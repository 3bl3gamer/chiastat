#!/bin/python3

import sys
import socket

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.connect(('127.0.0.1', 18445))
sockfile = sock.makefile()

while True:
    line = sockfile.readline().removesuffix('\n')
    print(line)
    cmd, host, port = line.split(' ')
    sock.send(f'H {"00"*32} {host} {port} 1.2.3.proto 2.3.4.soft FULL_NODE\n'.encode())
    sock.send(f'R {host} {port} {host} {port}\n'.encode())
