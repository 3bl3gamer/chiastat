import sys
import socket
from chia.protocols.protocol_message_types import ProtocolMessageTypes
from chia.protocols.shared_protocol import Handshake
from chia.protocols.full_node_protocol import RequestPeers
from chia.server.outbound_message import NodeType

def ep(*args):
    print(*args, file=sys.stderr)

sock_is_connected = False
def sock_connect():
    global sock
    global sock_is_connected
    ep('connecting to socket')
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect(('127.0.0.1', 18444))
    sock_is_connected = True
def send(msg, retry=True):
    global sock_is_connected
    if not sock_is_connected:
        sock_connect()
    try:
        sock.send((msg+'\n').encode())
    except OSError as ex:
        ep(ex)
        sock_is_connected = False
        if retry:
            send(msg, retry=False)

def hook(ws_conn):
    try:
        if not hasattr(ws_conn.__class__, '_no_hook'):
            hook_inner(ws_conn)
    except Exception as ex:
        ep(ex)

def hook_inner(ws):
    ep('hello from hook')
    ep('ws id', ws.peer_node_id, ws.peer_host)

    _read_one_message_orig = ws._read_one_message
    async def _read_one_message():
        msg = await _read_one_message_orig()
        try:
            if msg is not None and ProtocolMessageTypes(msg.type) == ProtocolMessageTypes.handshake:
                handshake = Handshake.from_bytes(msg.data)
                ep('handshake', handshake.protocol_version, 'is out:', ws.is_outbound)
                node_type_name = NodeType(handshake.node_type).name
                if not ws.is_outbound:
                    send(f'H {ws.peer_node_id} {ws.peer_host} {handshake.server_port} {handshake.protocol_version} {handshake.software_version} {node_type_name}')
        except Exception as ex:
            ep(ex)
        return msg
    ws._read_one_message = _read_one_message

    perform_handshake_orig = ws.perform_handshake
    async def perform_handshake(network_id, protocol_version, server_port, local_type):
        res = await perform_handshake_orig(network_id, protocol_version, server_port, local_type)
        try:
            ep('phs', network_id, protocol_version, 'is out:', ws.is_outbound, ws.local_type, '->', res.node_type)
            # for i in range(3):
            #     peers_resp = await ws.request_peers(RequestPeers())
            #     ep('peers:', peers_resp and len(peers_resp.peer_list))
            #     if peers_resp is not None:
            #         send(f'R {" ".join(f"{p.host} {p.port}" for p in peers_resp.peer_list)}')
        except Exception as ex:
            ep(ex)
        return res
    ws.perform_handshake = perform_handshake

    ep('hooked')
