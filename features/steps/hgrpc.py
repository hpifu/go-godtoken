#!/usr/bin/env python3

import grpc
import json
from hamcrest import *
from google.protobuf.json_format import MessageToJson


def check(obj1, obj2):
    if isinstance(obj2, dict):
        for key in obj2:
            check(obj1[key], obj2[key])
    elif isinstance(obj2, list):
        for idx, val in enumerate(obj2):
            check(obj1[idx], val)
    else:
        assert_that(obj1, equal_to(obj2))


@when('grpc 请求 {mod:str} {rpc:str}')
def step_impl(context, mod, rpc):
    req = json.loads(context.text)
    grpc_cli = getattr(__import__(mod + '_pb2_grpc'), 'ServiceStub')(context.grpc_chan)
    res = getattr(grpc_cli, rpc)(getattr(__import__(mod + '_pb2'), rpc + 'Req')(**req))
    context.res = json.loads(MessageToJson(res, including_default_value_fields=True))
    print(context.res)


@Then('grpc 检查')
def step_impl(context):
    res = json.loads(context.text)
    check(context.res, res)


# if __name__ == '__main__':
#     channel = grpc.insecure_channel('localhost:17060')
#     grpc_cli = getattr(__import__('godtoken' + '_pb2_grpc'), 'ServiceStub')(channel)
#     req = {
#         "rid": "123"
#     }
#     res = getattr(grpc_cli, 'GetToken')(getattr(__import__('godtoken' + '_pb2'), 'GetTokenReq')(**req))
#     print(res)
