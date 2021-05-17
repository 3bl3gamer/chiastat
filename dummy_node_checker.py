#!/bin/python3

import sys
import os
import asyncio

async def bg_print():
    while True:
        print('should not block')
        await asyncio.sleep(1)

async def main():
    asyncio.create_task(bg_print())

    f_req = os.open('update_nodes_request.fifo', os.O_RDONLY | os.O_NONBLOCK)
    f_res = os.open('update_nodes_response.fifo', os.O_WRONLY)
    remaining_chunk = b''
    packet_num = 0
    while True:
        try:
            chunk = os.read(f_req, 1024)
        except BlockingIOError:
            print('empty (ex)')
            await asyncio.sleep(1)
            continue
        if chunk == b'':
            print('empty')
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
            print(msg)
            cmd, host, port = msg.split(' ')
            os.write(f_res, f'H {packet_num} {"00"*32} {host} {port} 1.2.3.proto 2.3.4.soft FULL_NODE\n'.encode())
            packet_num += 1
            os.write(f_res, f'R {packet_num} {host} {port} {host} {port}\n'.encode())
            packet_num += 1

loop = asyncio.new_event_loop()
asyncio.set_event_loop(loop)
result = loop.run_until_complete(main())
