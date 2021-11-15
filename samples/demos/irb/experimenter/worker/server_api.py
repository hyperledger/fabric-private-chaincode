# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

import attestation
import keys
import protos.irb_pb2 as irb
import storage_client
import diagnose
import base64

def HandleAttestation():
    attestation_bytes = attestation.GetAttestation()
    print(attestation_bytes)

    #build identity bytes
    id = irb.Identity()
    id.public_key = keys.GetSerializedVerifyingKey()
    id.public_encryption_key = keys.GetSerializedEncryptionKey()
    id_bytes = id.SerializeToString()

    #build worker credentials
    wc = irb.WorkerCredentials()
    wc.identity_bytes = id_bytes
    wc.attestation = attestation_bytes.encode('utf_8')
    return wc.SerializeToString()

def HandleExecuteEvaluationPack(evaluation_pack):
    eep = irb.EncryptedEvaluationPack()
    eep.ParseFromString(evaluation_pack)

    #decrypt encrypted evalution pack
    res = keys.PkDecrypt(eep.encrypted_encryption_key)
    if res[1] != None:
        return "Error decrypting eval encryption key: " + res[1]
    key = bytearray(res[0])
    res = keys.Decrypt(key, eep.encrypted_evaluationpack)
    if res[1] != None:
        return "Error decrypting eval pack: " + res[1]
    evaluationpack = bytearray(res[0])

    epm = irb.EvaluationPackMessage()
    epm.ParseFromString(evaluationpack)

    sc = storage_client.StorageClient()

    return_string = ""

    first = True
    for r in epm.registered_data:
        if first:
            first = False
        else:
            return_string += ", "

        print("Data handler: {0}".format(r.data_handler))

        val = sc.get(r.data_handler)
        if val[1] == False:
            return "Error: Cannot retrieve value of key: " + r.data_handler

        print("Encoded data: {0}".format(val[0]))
        encrypted_data = base64.b64decode(val[0])

        res = keys.Decrypt(r.decryption_key,  encrypted_data)
        if res[1] != None:
            return "Error decrypting data item: " + res[1]
        data = bytearray(res[0])

        result = diagnose.diagnose(data)
        if result[1] != "" :
            return "Error running diagnosis: " + result[1]

        return_string += result[0]

    print("Final result: {0}".format(return_string))
    return return_string
