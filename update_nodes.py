#!/bin/python3

import sys
import os
work_dir = os.path.dirname(os.path.realpath(__file__))
sys.path.append(work_dir + '/chia-blockchain')

# https://github.com/Chia-Network/chia-blockchain/blob/9cc908678b1255c9a520c322302ba35084676e08/chia/server/server.py

import logging
import asyncio
import aiohttp
import ssl
import ipaddress
import io
import time
from cryptography import x509
from cryptography.hazmat.backends import default_backend
from cryptography.hazmat.primitives import hashes
from chia.types.blockchain_format.sized_bytes import bytes32
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional, Set, Tuple
from chia.server.ws_connection import WSChiaConnection
from chia.protocols.shared_protocol import protocol_version
from aiohttp import ClientSession, ClientTimeout, ServerDisconnectedError, client_exceptions
from chia.server.outbound_message import NodeType
from chia.protocols.shared_protocol import Handshake
from chia.protocols.protocol_message_types import ProtocolMessageTypes
from chia.protocols.full_node_protocol import RequestPeers


f_req = os.open('update_nodes_request.fifo', os.O_RDONLY | os.O_NONBLOCK)
f_res = os.open('update_nodes_response.fifo', os.O_WRONLY)
def send(msg):
    os.write(f_res, (msg+'\n').encode())

def ssl_context_for_client(
    ca_cert: Path,
    ca_key: Path,
    private_cert_path: Path,
    private_key_path: Path,
) -> Optional[ssl.SSLContext]:
    ssl_context = ssl._create_unverified_context(purpose=ssl.Purpose.SERVER_AUTH, cafile=str(ca_cert))
    ssl_context.check_hostname = False
    ssl_context.load_cert_chain(certfile=str(private_cert_path), keyfile=str(private_key_path))
    ssl_context.verify_mode = ssl.CERT_REQUIRED
    return ssl_context

async def get_info_from(target_host, target_port):
    timeout = ClientTimeout(total=2)
    session = ClientSession(timeout=timeout)
    try:
        return await get_info_from_inner(target_host, target_port, session)
    finally:
        await session.close()

peer_times = {}
async def get_info_from_inner(target_host, target_port, session):
    local_type = NodeType.FULL_NODE
    self_port = 8444

    incoming_messages: asyncio.Queue = asyncio.Queue()

    home = str(Path.home())
    ssl_context = ssl_context_for_client(
        home + "/.chia/mainnet/config/ssl/ca/chia_ca.crt",
        home + "/.chia/mainnet/config/ssl/ca/chia_ca.key",
        home + "/.chia/mainnet/config/ssl/full_node/public_full_node.crt",
        home + "/.chia/mainnet/config/ssl/full_node/public_full_node.key"
    )

    try:
        ipaddress.IPv6Address(target_host)
        target_host_ = f'[{target_host}]'
    except ipaddress.AddressValueError:
        target_host_ = target_host
    url = f"wss://{target_host_}:{target_port}/ws"
    ws = await session.ws_connect(
        url, autoclose=True, autoping=True, heartbeat=60, ssl=ssl_context, max_msg_size=50 * 1024 * 1024
    )

    transport = ws._response.connection.transport  # type: ignore
    cert_bytes = transport._ssl_protocol._extra["ssl_object"].getpeercert(True)  # type: ignore
    der_cert = x509.load_der_x509_certificate(cert_bytes, default_backend())
    peer_id = bytes32(der_cert.fingerprint(hashes.SHA256()))

    WSChiaConnection._no_hook = True
    connection = WSChiaConnection(
        local_type,
        ws,
        self_port,
        logging.getLogger(__name__),
        True,
        False,
        target_host,
        incoming_messages,
        lambda a,b: None,
        peer_id,
        100,
        30,
        session=session,
    )

    f = connection._read_one_message
    handshake = None
    async def _read_one_message():
        nonlocal handshake
        msg = await f()
        if msg is not None and ProtocolMessageTypes(msg.type) == ProtocolMessageTypes.handshake:
            handshake = Handshake.from_bytes(msg.data)
        return msg
    connection._read_one_message = _read_one_message

    handshake_res = await connection.perform_handshake(
        'mainnet',
        protocol_version,
        self_port,
        local_type,
    )
    assert handshake_res is True

    all_peers = []
    for iter in range(3):
        peers_resp = await connection.request_peers(RequestPeers())
        if peers_resp is None:
            break
        else:
            all_peers.extend(peers_resp.peer_list)

    # print(f'{target_host}, {len(all_peers)}')
    peers = []
    now = time.time()
    thresh = now - 5*60
    for peer in all_peers:
        key = f'{peer.host} {peer.port}'
        stamp = peer_times.get(key)
        if stamp is None or stamp < thresh:
            peers.append(peer)
            peer_times[key] = now
    print(f'using peers: {len(peers)}/{len(all_peers)}')
    return (peer_id, handshake, peers)

count_total = 0
count_ok = 0
async def try_get_info_from(host, port):
    global count_total
    global count_ok
    try:
        res = await get_info_from(host, port)
        count_ok += 1
        return res
    except (asyncio.TimeoutError, aiohttp.client_exceptions.ClientConnectorError, aiohttp.client_exceptions.ServerDisconnectedError, ConnectionResetError) as ex:
        # print(f'fail {host}:{port}', ex, type(ex))
        return None
    finally:
        count_total += 1
        if count_total % 100 == 0:
            print(f'checked {count_total} nodes, {count_ok} ok')

async def checking_worker(in_queue, out_queue):
    while True:
        try:
            host, port = await in_queue.get()
            id_hs_peers = await try_get_info_from(host, port)
            if id_hs_peers is not None:
                peer_id, hs, peers = id_hs_peers
                await out_queue.put((peer_id, host, port, hs))
                await out_queue.put((peers,))
            in_queue.task_done()
        except Exception as ex:
            print('checking_worker', ex)

async def sending_worker(queue):
    packet_num = 0
    while True:
        try:
            item = await queue.get()
            if len(item) == 4:
                node_id, host, port, handshake = item
                node_type_name = NodeType(handshake.node_type).name
                send(f'H {packet_num} {node_id} {host} {port} {handshake.protocol_version} {handshake.software_version} {node_type_name}')
                packet_num += 1
            else:
                peers = item[0]
                if len(peers) > 0:
                    send(f'R {packet_num} {" ".join(f"{p.host} {p.port}" for p in peers)}')
                    packet_num += 1
            queue.task_done()
        except Exception as ex:
            print('sending_worker', ex)

async def queue_sizes_worker(in_queue, out_queue):
    while True:
        print(f'queues: in={in_queue.qsize()}, out={out_queue.qsize()}')
        await asyncio.sleep(5)

async def main():
    task_count = 128
    old_addrs_queue = asyncio.Queue(maxsize=128)
    new_data_queue = asyncio.Queue(maxsize=128)

    [asyncio.create_task(checking_worker(old_addrs_queue, new_data_queue)) for i in range(task_count)]
    asyncio.create_task(sending_worker(new_data_queue))
    asyncio.create_task(queue_sizes_worker(old_addrs_queue, new_data_queue))

    remaining_chunk = b''
    while True:
        try:
            chunk = os.read(f_req, 1024)
        except BlockingIOError:
            await asyncio.sleep(1)
            continue
        if chunk == b'':
            await asyncio.sleep(1)
            continue

        chunk = remaining_chunk + chunk
        remaining_chunk = b''
        pos = 0
        while True:
            end_pos = chunk.find(b'\n', pos)
            if end_pos == -1:
                remaining_chunk = chunk[pos:]
                break
            msg = chunk[pos:end_pos].decode()
            pos = end_pos + 1
            if msg[0] == 'C':
                try:
                    _, host, port = msg.split(' ')
                except ValueError:
                    raise ValueError(f'wrong message: {msg}')
                else:
                    await old_addrs_queue.put((host, int(port)))
            else:
                raise ValueError(f'unexpected message ({msg[0]}): {msg}')

loop = asyncio.new_event_loop()
asyncio.set_event_loop(loop)
result = loop.run_until_complete(main())
