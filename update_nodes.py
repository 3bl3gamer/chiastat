import sys
import os
work_dir = os.path.dirname(os.path.realpath(__file__))
sys.path.append(work_dir + '/chia-blockchain')

# https://github.com/Chia-Network/chia-blockchain/blob/9cc908678b1255c9a520c322302ba35084676e08/chia/server/server.py

import socket
import logging
import asyncio
import aiohttp
import ssl
import ipaddress
import io
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


sock_is_connected = False
def sock_connect():
    global sock
    global sock_is_connected
    print('connecting to socket')
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect(('127.0.0.1', 18445))
    sock.setblocking(False)
    sock_is_connected = True
def send(msg, retry=True):
    global sock_is_connected
    if not sock_is_connected:
        sock_connect()
    try:
        sock.send((msg+'\n').encode())
    except OSError as ex:
        print(ex)
        sock_is_connected = False
        if retry:
            send(msg, retry=False)


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
        if ProtocolMessageTypes(msg.type) == ProtocolMessageTypes.handshake:
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

    peers_resp = await connection.request_peers(RequestPeers())
    if peers_resp is None:
        peers = []
    else:
        peers = peers_resp.peer_list
    return (peer_id, handshake, peers)

async def try_get_info_from(host, port):
    try:
        return await get_info_from(host, port)
    except (asyncio.TimeoutError, aiohttp.client_exceptions.ClientConnectorError, aiohttp.client_exceptions.ServerDisconnectedError):
        print(f'fail  {host}:{port}')
        return None

async def checking_worker(in_queue, out_queue):
    while True:
        host, port = await in_queue.get()
        for iter in range(3):
            id_hs_peers = await try_get_info_from(host, port)
            if id_hs_peers is None:
                break
            peer_id, hs, peers = id_hs_peers
            if iter == 0:
                await out_queue.put((peer_id, host, port, hs))
            await out_queue.put((peers,))
            if len(peers) == 0:
                break
        in_queue.task_done()

async def sending_worker(queue):
    while True:
        item = await queue.get()
        if len(item) == 4:
            node_id, host, port, handshake = item
            node_type_name = NodeType(handshake.node_type).name
            send(f'H {node_id} {host} {port} {handshake.protocol_version} {handshake.software_version} {node_type_name}')
        else:
            peers = item[0]
            send(f'R {" ".join(f"{p.host} {p.port}" for p in peers)}')
        queue.task_done()

async def main():
    global sock_is_connected

    task_count = 128
    old_addrs_queue = asyncio.Queue(maxsize=128)
    new_data_queue = asyncio.Queue(maxsize=128)

    check_tasks = [asyncio.create_task(checking_worker(old_addrs_queue, new_data_queue)) for i in range(task_count)]
    out_task = asyncio.create_task(sending_worker(new_data_queue))

    while True:
        print(f'iter: {old_addrs_queue.qsize()}/{old_addrs_queue.maxsize} {new_data_queue.qsize()}/{new_data_queue.maxsize}')

        with io.BytesIO() as buffer:
            while True:
                try:
                    if not sock_is_connected:
                        sock_connect()
                    resp = sock.recv(100)
                except ConnectionRefusedError:
                    print('socket connection refused')
                    buffer.truncate(0)
                    await asyncio.sleep(5)
                except BlockingIOError:
                    await asyncio.sleep(1)
                else:
                    if len(resp) == 0:
                        print('socket seems disconnected')
                        sock_is_connected = False
                        continue
                    buffer.write(resp)
                    buffer.seek(0)
                    start_index = 0
                    for line in buffer:
                        start_index += len(line)
                        msg = line.decode().removesuffix('\n')
                        if msg[0] == 'C':
                            _, host, port = msg.split(' ')
                            await old_addrs_queue.put((host, int(port)))
                        else:
                            raise ValueError('unexpected message: ' + msg)

                    if start_index:
                        buffer.seek(start_index)
                        remaining = buffer.read()
                        buffer.truncate(0)
                        buffer.seek(0)
                        buffer.write(remaining)
                    else:
                        buffer.seek(0, 2)

loop = asyncio.new_event_loop()
asyncio.set_event_loop(loop)
result = loop.run_until_complete(main())
