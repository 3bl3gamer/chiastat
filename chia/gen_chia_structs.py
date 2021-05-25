#!/bin/python3

import os
import ast
import urllib.request
from textwrap import dedent


DEBUG = False
workdir = os.path.dirname(os.path.realpath(__file__))

def cap_first(string):
    if len(string) == 0:
        return string
    return string[0].upper() + string[1:]

def to_camel_case(string):
    items = string.split('_')
    return items[0] + ''.join(cap_first(x) for x in items[1:])

def to_pascal_case(string):
    return ''.join(cap_first(x) for x in string.split('_'))


def get_annotation(annotation):
    if isinstance(annotation, ast.Name):
        return [annotation.id]
    elif isinstance(annotation, ast.Subscript):
        return [annotation.value.id] + get_annotation(annotation.slice)
    else:
        raise ValueError(f'unexpected annotation')

def is_class_type_name(name):
    return name != 'List' and name != 'Optional' and name[0] == name[0].upper()

def make_type_def(ann_items):
    if len(ann_items) == 0:
        return ''
    def get_next_def():
        t = ann_items[0]
        if t == 'bool':
            return 'bool'
        if t == 'uint8':
            return 'uint8'
        if t == 'uint32':
            return 'uint32'
        if t == 'uint64':
            return 'uint64'
        if t == 'uint128':
            return '*big.Int'
        if t == 'bytes32':
            return '[32]byte'
        if t == 'bytes100':
            return '[100]byte'
        if t == 'bytes':
            return '[]byte'
        if t == 'List':
            return '[]'
        if t == 'Optional':
            return '*' if is_class_type_name(ann_items[1]) else ''
        if is_class_type_name(t):
            return t
        raise ValueError(f'unexpected type {t} in {ann_items}')
    return get_next_def() + make_type_def(ann_items[1:])

build_in_parse_funcs = {
    'BoolFromBytes',
    'Uint8FromBytes',
    'Uint32FromBytes',
    'Uint64FromBytes',
    'Uint128FromBytes',
    'Bytes32FromBytes',
    'Bytes100FromBytes',
    'BytesFromBytes',
}

def make_type_parse(name, ann_items, need_err_check=False):
    func_name = ''.join([cap_first(x) for x in ann_items]) + 'FromBytes'
    res = ''
    if func_name in build_in_parse_funcs:
        res += f'{name} = {func_name}(buf)\n'
    elif len(ann_items) == 1 and is_class_type_name(ann_items[0]):
        res += f'{name} = {func_name}(buf)\n'
    elif ann_items[0] == 'Optional':
        inner_type_def = make_type_def(ann_items[1:])
        res += f'if flag := BoolFromBytes(buf); buf.err == nil && flag {{ \n'
        if is_class_type_name(ann_items[1]):
            res += f'{make_type_parse("var t", ann_items[1:])}'
            res += f'{name} = &t\n'
        else:
            res += f'{make_type_parse(name, ann_items[1:])}'
        res += f'}}\n'
    elif ann_items[0] == 'List':
        inner_type_def = make_type_def(ann_items[1:])
        len_name = 'len_' + name.replace('.', '_')
        res += f'{len_name} := Uint32FromBytes(buf)\n'
        if DEBUG:
            res += f'fmt.Println("len", {len_name})\n'
        res += f'{name} = make([]{inner_type_def}, {len_name})\n'
        res += f'for i:=uint32(0);i<{len_name};i++ {{ {make_type_parse(f"{name}[i]", ann_items[1:], need_err_check=True)} }}\n'
    if DEBUG:
        res += f'fmt.Println("{name}", "{make_type_def(ann_items)}", {name}, buf.pos, buf.err)\n'
    if need_err_check:
        res += f'if buf.err != nil {{ return }}\n'
    return res

def make_struct_def(modules, struct_name):
    has_found = False
    docstring = None
    attrs = []

    for (source_lines, module) in modules:
        for item in module.body:
            if isinstance(item, ast.ClassDef) and item.name == struct_name:
                has_found = True
                if len(item.body) > 0 and isinstance(item.body[0], ast.Expr) and isinstance(item.body[0].value, ast.Constant):
                    docstring = item.body[0].value.value
                for attr in item.body:
                    if isinstance(attr, ast.AnnAssign):
                        attr_docstring = None
                        line_end = source_lines[attr.annotation.end_lineno-1][attr.annotation.end_col_offset:].strip()
                        if line_end != '':
                            attr_docstring = line_end.removeprefix('#').strip()
                        ann_items = get_annotation(attr.annotation)
                        if ann_items[0] == 'Optional':
                            attr_docstring = '(optional)' + ('' if attr_docstring is None else ' ' + attr_docstring)
                        attrs.append((attr.target.id, ann_items, attr_docstring))


    if not has_found:
        raise ValueError(f'{struct_name} definition was not found')

    def_text = ''
    if docstring is not None:
        text = dedent(docstring).removeprefix('\n').removesuffix('\n')
        def_text += '// ' + text.replace('\n', '\n// ') + '\n'
    def_text += f'type {struct_name} struct ' + '{\n'
    for (name, ann_items, attr_docstring) in attrs:
        if attr_docstring is not None:
            def_text += '// ' + attr_docstring + '\n'
        def_text += to_pascal_case(name) + ' ' + make_type_def(ann_items) + '\n'
    def_text += '}\n'

    parse_text = f'func {struct_name}FromBytes(buf *ParseBuf) (obj {struct_name}) {{\n'
    for (name, ann_items, attr_docstring) in attrs:
        parse_text += make_type_parse('obj.' + to_pascal_case(name), ann_items)
    parse_text += f'return\n'
    parse_text += '}\n'

    return def_text + '\n\n' + parse_text

source_urls = [
    'https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/consensus/block_record.py',
    'https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/types/blockchain_format/coin.py',
    'https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/types/blockchain_format/classgroup.py',
    'https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/types/blockchain_format/sub_epoch_summary.py',
    'https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/types/blockchain_format/vdf.py',
    'https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/types/blockchain_format/foliage.py',
    'https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/types/blockchain_format/reward_chain_block.py',
    'https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/types/blockchain_format/program.py',
    'https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/types/blockchain_format/slots.py',
    'https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/types/blockchain_format/pool_target.py',
    'https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/types/blockchain_format/proof_of_space.py',
    'https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/types/full_block.py',
    'https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/types/end_of_slot_bundle.py',
]

modules = []
for url in source_urls:
    text = urllib.request.urlopen(url).read().decode('utf-8')
    modules.append((text.split('\n'), ast.parse(text)))

out_fname = workdir + '/structs_generated.go'

with open(out_fname, 'w') as f:
    imports = ["math/big"]
    if DEBUG:
        imports.append("fmt")
    f.write('package chia\n\n')
    f.write('import (\n' + '\n'.join(f'"{x}"' for x in imports) + '\n)\n\n')
    f.write('\n\n'.join(make_struct_def(modules, name) for name in [
        'BlockRecord', 'Coin', 'ClassgroupElement', 'SubEpochSummary',
        'FullBlock', 'EndOfSubSlotBundle', 'VDFProof', 'VDFInfo',
        'Foliage', 'FoliageTransactionBlock', 'FoliageBlockData', 'TransactionsInfo', 'RewardChainBlock',
        'ChallengeChainSubSlot', 'InfusedChallengeChainSubSlot', 'RewardChainSubSlot', 'SubSlotProofs',
        'PoolTarget', 'ProofOfSpace'
    ]))
    for dummy_name, dummy_size in [('G1Element', 48), ('G2Element', 96)]:
        f.write(f'\n\ntype {dummy_name} struct {{Bytes []byte}}')
        f.write(f'\nfunc {dummy_name}FromBytes(buf *ParseBuf) (obj {dummy_name}) {{')
        f.write(f'\nobj.Bytes = BytesNFromBytes(buf, {dummy_size}); return')
        f.write(f'\n}}')

os.system('go fmt ' + out_fname)
