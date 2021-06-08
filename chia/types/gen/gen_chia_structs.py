#!/bin/python3

import os
import ast
import urllib.request
from textwrap import dedent


DEBUG = False
workdir = os.path.dirname(os.path.realpath(__file__)) + "/.."

def cap_first(string):
    if len(string) == 0:
        return string
    return string[0].upper() + string[1:]

def to_camel_case(string):
    items = string.split('_')
    return items[0] + ''.join(cap_first(x) for x in items[1:])

def to_pascal_case(string):
    return ''.join(cap_first(x) for x in string.split('_'))

def to_attr_name(string):
    name = to_pascal_case(string)
    if name.endswith('Id'):
        name = name[:-2] + 'ID'
    return name


def get_annotation(annotation):
    if isinstance(annotation, ast.Name):
        return [annotation.id]
    elif isinstance(annotation, ast.Subscript):
        if annotation.value.id == 'Tuple':
            return get_annotation(annotation.slice)  # do not need string 'Tuple' in list
        return [annotation.value.id] + get_annotation(annotation.slice)
    elif isinstance(annotation, ast.Tuple):
        return [('Tuple',) + tuple(get_annotation(dim) for dim in annotation.dims)]
    else:
        raise ValueError(f'unexpected annotation {annotation}')

def is_class_type_name(name):
    return isinstance(name, str) and name != 'List' and name != 'Optional' and name[0] == name[0].upper()

def type_option_is_ref(name):
    return is_class_type_name(name) or (name != 'bytes' and name.startswith('bytes'))

def make_tuple_struct_name(tuple_ann_items):
    return 'Tuple' + ''.join(''.join(cap_first(x) for x in dim) for dim in tuple_ann_items[1:])

def make_type_def(ann_items):
    if len(ann_items) == 0:
        return ''
    def get_next_def():
        t = ann_items[0]
        if t == 'bool':
            return 'bool'
        if t == 'uint8':
            return 'uint8'
        if t == 'uint16':
            return 'uint16'
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
        if t == 'str':
            return 'string'
        if t == 'List':
            return '[]'
        if isinstance(t, tuple) and t[0] == 'Tuple':
            return make_tuple_struct_name(t)
        if t == 'Optional':
            return '*' if type_option_is_ref(ann_items[1]) else ''
        if is_class_type_name(t):
            return t
        raise ValueError(f'unexpected type {t} in {ann_items}')
    return get_next_def() + make_type_def(ann_items[1:])

def make_type_name(ann_item):
    if ann_item == 'str':
        return 'string'
    if isinstance(ann_item, tuple) and ann_item[0] == 'Tuple':
        return make_tuple_struct_name(ann_item)
    return ann_item

buf_func_names = {
    'Bool',
    'Uint8',
    'Uint16',
    'Uint32',
    'Uint64',
    'Uint128',
    'Bytes32',
    'Bytes100',
    'Bytes',
    'String'
}

build_in_parse_funcs = {
    'BoolFromBytes',
    'Uint8FromBytes',
    'Uint16FromBytes',
    'Uint32FromBytes',
    'Uint64FromBytes',
    'Uint128FromBytes',
    'Bytes32FromBytes',
    'Bytes100FromBytes',
    'BytesFromBytes',
    'StringFromBytes',
}

def make_parse_func_call(attr_name, ann_items):
    base = ''.join([cap_first(make_type_name(x)) for x in ann_items])
    if base in buf_func_names:
        return f'{attr_name} = buf.{base}()'
    return f'{attr_name}.FromBytes(buf)'

def make_serialize_func_call(attr_name, ann_items):
    base = ''.join([cap_first(make_type_name(x)) for x in ann_items])
    if base in buf_func_names:
        return f'utils.{base}ToBytes(buf, {attr_name})'
    return f'{attr_name}.ToBytes(buf)'

def make_none_option_check(attr_name, ann_items):
    t = ann_items[1]
    if t.startswith('uint'):
        return f'{attr_name} == 0'
    elif t == 'List':
        return f'len({attr_name}) == 0'
    elif type_option_is_ref(t):
        return f'{attr_name} == nil'
    else:
        raise ValueError(f'serializing optional {t} is not supported ({attr_name})')

def get_tuples_from(ann_items):
    for item in ann_items:
        if isinstance(item, tuple) and item[0] == 'Tuple':
            yield item

def make_type_parse(name, ann_items, need_err_check=False):
    res = ''
    if len(ann_items) == 1:
        res += make_parse_func_call(name, ann_items) + '\n'
    elif ann_items[0] == 'Optional':
        res += f'if flag := buf.Bool(); buf.Err() == nil && flag {{ \n'
        if type_option_is_ref(ann_items[1]):
            res += f'var t {make_type_def(ann_items[1:])}\n'
            res += f'{make_type_parse("t", ann_items[1:])}'
            res += f'{name} = &t\n'
        else:
            res += f'{make_type_parse(name, ann_items[1:])}'
        res += f'}}\n'
    elif ann_items[0] == 'List':
        inner_type_def = make_type_def(ann_items[1:])
        len_name = 'len_' + name.replace('.', '_')
        res += f'{len_name} := buf.Uint32()\n'
        if DEBUG:
            res += f'fmt.Println("len", {len_name})\n'
        res += f'{name} = make([]{inner_type_def}, {len_name})\n'
        res += f'for i:=uint32(0);i<{len_name};i++ {{ {make_type_parse(f"{name}[i]", ann_items[1:], need_err_check=True)} }}\n'
    if DEBUG:
        res += f'fmt.Println("{name}", "{make_type_def(ann_items)}", {name}, buf.Pos(), buf.Err())\n'
    if need_err_check:
        res += f'if buf.Err() != nil {{ return }}\n'
    return res

def make_type_serialize(name, ann_items):
    res = ''
    if len(ann_items) == 1:
        res += make_serialize_func_call(name, ann_items) + '\n'
    elif ann_items[0] == 'Optional':
        opt_name = name.replace('.', '_') + '_isSet'
        deref = '*' if type_option_is_ref(ann_items[1]) and not is_class_type_name(ann_items[1]) else ''
        res += f'{opt_name} := !({make_none_option_check(name, ann_items)})\n'
        res += f'utils.BoolToBytes(buf, {opt_name})\n'
        res += f'if {opt_name} {{\n'
        res += make_type_serialize(deref+name, ann_items[1:])
        res += '}\n'
    elif ann_items[0] == 'List':
        res += f'utils.Uint32ToBytes(buf, uint32(len({name})))\n'
        res += f'for _, item := range {name} {{\n'
        res += make_type_serialize('item', ann_items[1:])
        res += '}\n'
    return res

def extract_struct_def_data(source_lines, module_ast, struct_name):
    was_found = False
    docstring = None
    attrs = []
    tuples = []

    for item in module_ast.body:
        if isinstance(item, ast.ClassDef) and item.name == struct_name:
            was_found = True
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
                    tuples.extend(get_tuples_from(ann_items))

    if not was_found:
        raise ValueError(f'{struct_name} definition was not found')

    return {'struct_name': struct_name, 'docstring': docstring, 'attrs': attrs, 'tuples': tuples}

def make_struct_def(struct_def_data):
    struct_name = struct_def_data['struct_name']
    docstring = struct_def_data['docstring']
    attrs = struct_def_data['attrs']

    # type def
    def_text = ''
    if docstring is not None:
        text = dedent(docstring).removeprefix('\n').removesuffix('\n')
        def_text += '// ' + text.replace('\n', '\n// ') + '\n'
    def_text += f'type {struct_name} struct ' + '{\n'
    for (name, ann_items, attr_docstring) in attrs:
        if attr_docstring is not None:
            def_text += '// ' + attr_docstring + '\n'
        def_text += to_attr_name(name) + ' ' + make_type_def(ann_items) + '\n'
    def_text += '}\n'

    # from bytes
    parse_text = f'func (obj *{struct_name}) FromBytes(buf *utils.ParseBuf) {{\n'
    for (name, ann_items, attr_docstring) in attrs:
        parse_text += make_type_parse('obj.' + to_attr_name(name), ann_items)
    parse_text += '}\n'

    # to bytes
    ser_text = f'func (obj {struct_name}) ToBytes(buf *[]byte) {{\n'
    for (name, ann_items, attr_docstring) in attrs:
        ser_text += make_type_serialize('obj.' + to_attr_name(name), ann_items)
    ser_text += '}\n'

    return def_text + '\n\n' + parse_text + '\n\n' + ser_text

def make_tuple_def(tup_items):
    name = make_tuple_struct_name(tup_items)
    attrs = [(f'v{i}', ann_items, None) for i, ann_items in enumerate(tup_items[1:])]
    return make_struct_def({'struct_name': name, 'docstring': None, 'attrs': attrs})


source_classes = {
    'consensus/block_record.py': ['BlockRecord'],
    'types/blockchain_format/coin.py': ['Coin'],
    'types/blockchain_format/classgroup.py': ['ClassgroupElement'],
    'types/blockchain_format/sub_epoch_summary.py': ['SubEpochSummary'],
    'types/blockchain_format/vdf.py': ['VDFProof', 'VDFInfo'],
    'types/blockchain_format/foliage.py': ['Foliage', 'FoliageTransactionBlock', 'FoliageBlockData', 'TransactionsInfo'],
    'types/blockchain_format/reward_chain_block.py': ['RewardChainBlock'],
    'types/blockchain_format/slots.py': ['ChallengeChainSubSlot', 'InfusedChallengeChainSubSlot', 'RewardChainSubSlot', 'SubSlotProofs'],
    'types/blockchain_format/pool_target.py': ['PoolTarget'],
    'types/blockchain_format/proof_of_space.py': ['ProofOfSpace'],
    'types/full_block.py': ['FullBlock'],
    'types/end_of_slot_bundle.py': ['EndOfSubSlotBundle'],
    # ===
    'server/outbound_message.py': ['Message'],
    'protocols/shared_protocol.py': ['Handshake'],
    'protocols/full_node_protocol.py': [
        'NewPeak',
        # 'NewTransaction', 'RequestTransaction', 'RespondTransaction',
        # 'RequestProofOfWeight', 'RespondProofOfWeight',
        # 'RequestBlock', 'RejectBlock', 'RequestBlocks', 'RespondBlocks', 'RejectBlocks', 'RespondBlock',
        # 'NewUnfinishedBlock', 'RequestUnfinishedBlock', 'RespondUnfinishedBlock',
        # 'NewSignagePointOrEndOfSubSlot', 'RequestSignagePointOrEndOfSubSlot', 'RespondSignagePoint', 'RespondEndOfSubSlot',
        'RequestMempoolTransactions',
        # 'NewCompactVDF', 'RequestCompactVDF', 'RespondCompactVDF',
        'RequestPeers', 'RespondPeers',
    ],
    'types/peer_info.py': ['TimestampedPeerInfo'],
}

modules = []
for path, classes in source_classes.items():
    url = 'https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/' + path
    text = urllib.request.urlopen(url).read().decode('utf-8')
    modules.append({
        'source_lines': text.split('\n'),
        'ast': ast.parse(text),
        'class_names': classes
    })

out_fname = workdir + '/structs_generated.go'

with open(out_fname, 'w') as f:
    imports = ["math/big", "chiastat/chia/utils"]
    if DEBUG:
        imports.append("fmt")
    f.write('// Generated, do not edit.\n')
    f.write('package types\n\n')
    f.write('import (\n' + '\n'.join(f'"{x}"' for x in imports) + '\n)\n\n')

    tuples = []
    for module in modules:
        for class_name in module['class_names']:
            data = extract_struct_def_data(module['source_lines'], module['ast'], class_name)
            f.write(make_struct_def(data) + '\n\n')
            tuples.extend(data['tuples'])

    f.write(f'\n\n// === Tuples ===\n\n')
    processed_tuple_names = set()
    for tup in tuples:
        name = make_tuple_struct_name(tup)
        if name not in processed_tuple_names:
            f.write(make_tuple_def(tup) + '\n\n')
            processed_tuple_names.add(name)

    f.write(f'\n\n// === Dummy ===\n\n')
    for dummy_name, dummy_size in [('G1Element', 48), ('G2Element', 96)]:
        f.write(f'\ntype {dummy_name} struct {{Bytes []byte}}')
        f.write(f'\nfunc (obj *{dummy_name}) FromBytes(buf *utils.ParseBuf) {{')
        f.write(f'\nobj.Bytes = buf.BytesN({dummy_size})')
        f.write(f'\n}}')
        f.write(f'\nfunc (obj {dummy_name}) ToBytes(buf *[]byte) {{')
        f.write(f'\nutils.BytesWOSizeToBytes(buf, obj.Bytes)')
        f.write(f'}}\n')


os.system('go fmt ' + out_fname)
