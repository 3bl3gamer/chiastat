import sys
import socket
from chia.protocols.protocol_message_types import ProtocolMessageTypes
from chia.protocols.shared_protocol import Handshake
from chia.protocols.full_node_protocol import RequestPeers
from chia.server.outbound_message import NodeType

def ep(*args):
    print(*args, file=sys.stderr)

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
sock_addr = ('127.0.0.1', 18444)
def send(msg):
    sock.sendto((msg+'\n').encode(), sock_addr)

def hook(ws_conn):
    try:
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
                ep('handshake', handshake.protocol_version)
                node_type_name = NodeType(handshake.node_type).name
                send(f'H {ws.peer_node_id} {ws.peer_host} {ws.peer_port} {handshake.protocol_version} {handshake.software_version} {node_type_name}')
        except Exception as ex:
            ep(ex)
        return msg
    ws._read_one_message = _read_one_message

    perform_handshake_orig = ws.perform_handshake
    async def perform_handshake(network_id, protocol_version, server_port, local_type):
        res = await perform_handshake_orig(network_id, protocol_version, server_port, local_type)
        try:
            ep('phs', network_id, protocol_version, 'is out', ws.is_outbound, ws.local_type)
            for i in range(5):
                peers_resp = await ws.request_peers(RequestPeers())
                ep(peers_resp and len(peers_resp.peer_list))
                if peers_resp is not None:
                    for peer in peers_resp.peer_list:
                        send(f'R {peer.host} {peer.port}')
        except Exception as ex:
            ep(ex)
        return res
    ws.perform_handshake = perform_handshake

    ep('hooked')
