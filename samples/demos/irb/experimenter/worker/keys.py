from cryptography.fernet import Fernet
from cryptography.hazmat.primitives.asymmetric import ec
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
import logging

logger = logging.getLogger()

def InitializeKeys():
    global signing_key
    global verifying_key

    global decryption_key
    global encryption_key

    try:
        signing_key = ec.generate_private_key(ec.SECP256R1())
        verifying_key = signing_key.public_key()
    except Exception:
        logger.error("error creating signing keys")
        return False

    try:
        decryption_key = rsa.generate_private_key(public_exponent=65537,key_size=3072)
        encryption_key = decryption_key.public_key()
    except Exception:
        logger.error("error creating encryption keys")
        return False

    logger.info("Created keys:")
    logger.info("Verifying key: {0}".format(GetSerializedVerifyingKey()))
    logger.info("Encryption key: {0}".format(GetSerializedEncryptionKey()))

    return True

def GetSerializedVerifyingKey():
    return verifying_key.public_bytes(encoding=serialization.Encoding.PEM, format=serialization.PublicFormat.SubjectPublicKeyInfo)

def GetSerializedEncryptionKey():
    return encryption_key.public_bytes(encoding=serialization.Encoding.PEM, format=serialization.PublicFormat.PKCS1)

#This function decrypts the message with the private decryption key and returns m, err (Go style)
def PkDecrypt(encryptedMessageBytes):
    try:
        m = decryption_key.decrypt(
            encryptedMessageBytes,
            padding.OAEP(
                mgf=padding.MGF1(algorithm=hashes.SHA1()),
                algorithm=hashes.SHA1(),
                label=None
            ))

    except Exception as e:
        return None, str(e)

    return m, None

#This function decrypts the message with the symmetric key "key", and returns m, err (Go style)
def Decrypt(key, encryptedMessageBytes):
    NonceLength = 12
    TagLength = 16

    nonce = encryptedMessageBytes[:NonceLength]
    tag = encryptedMessageBytes[NonceLength : NonceLength+TagLength]
    ciphertext = encryptedMessageBytes[NonceLength+TagLength:]

    try:
        aesgcm = AESGCM(key)
        m = aesgcm.decrypt(nonce, ciphertext + tag, None)
    except Exception as e:
        return None, str(e)

    return m, None

