# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

import attestation
import keys
import protos.irb_pb2 as irb
import pkg.storage.py.storage_client as storage_client
import pdo.common.crypto as crypto
import diagnose
import base64

def HandleAttestation():
    attestation_bytes = attestation.GetAttestation()
    print(attestation_bytes)

    #build identity bytes
    id = irb.Identity()
    id.public_key = keys.GetVerifyingKey().Serialize().encode('utf_8')
    id.public_encryption_key = keys.GetEncryptionKey().Serialize().encode('utf_8')
    id_bytes = id.SerializeToString()

    #build worker credentials
    wc = irb.WorkerCredentials()
    wc.identity_bytes = id_bytes
    wc.attestation = attestation_bytes.encode('utf_8')
    return wc.SerializeToString()

def HandleExecuteEvaluationPack(evaluation_pack):
    eep = irb.EncryptedEvaluationPack()
    eep.ParseFromString(evaluation_pack)

    #TODO decrypt

    epm = irb.EvaluationPackMessage()
    epm.ParseFromString(eep.encrypted_evaluationpack)

    sc = storage_client.StorageClient()

    return_string = ""

    first = True
    for r in epm.registered_data:
        if first:
            first = False
        else:
            return_string += ", "

        print("Decryption key: {0}".format(r.decryption_key))
        print("Data handler: {0}".format(r.data_handler))

        val = sc.get(r.data_handler)
        if val[1] == False:
            return "Cannot retrieve value of key: " + r.data_handler

        print("Encoded data: {0}".format(val[0]))
        encrypted_data = base64.b64decode(val[0])

        data = bytearray(crypto.SKENC_DecryptMessage(r.decryption_key, encrypted_data))
        print("Retrieved data: {0}".format(data))

        result = diagnose.diagnose(data)
        if result[1] != "" :
            return "Error: " + result[1]

        return_string += result[0]

    print("Final result: {0}".format(return_string))
    return return_string
