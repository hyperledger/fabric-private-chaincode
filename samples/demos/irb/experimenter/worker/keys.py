import pdo.common.crypto as crypto
import logging

logger = logging.getLogger()

def InitializeKeys():
    global signing_key
    global verifying_key

    global decryption_key
    global encryption_key

    try:
        signing_key = crypto.SIG_PrivateKey()
        signing_key.Generate()
        verifying_key = signing_key.GetPublicKey()
    except Exception:
        logger.error("error creating signing keys")
        return False

    try:
        decryption_key = crypto.PKENC_PrivateKey()
        decryption_key.Generate()
        encryption_key = decryption_key.GetPublicKey()
    except Exception:
        logger.error("error creating encryption keys")
        return False

    logger.info("Created keys:")
    logger.info("Verifying key: {0}".format(verifying_key.Serialize()))
    logger.info("Encryption key: {0}".format(encryption_key.Serialize()))
    return True

def GetVerifyingKey():
    return verifying_key

def GetEncryptionKey():
    return encryption_key

#This function decrypts the message with the private decryption key and returns m, err (Go style)
def PkDecrypt(encryptedMessageBytes):
    try:
        m = decryption_key.DecryptMessage(encryptedMessageBytes)

    except Exception as e:
        return None, str(e)

    return m, None

#This function decrypts the message with the symmetric key "key", and returns m, err (Go style)
def Decrypt(key, encryptedMessageBytes):
    try:
        m = crypto.SKENC_DecryptMessage(key, encryptedMessageBytes)

    except Exception as e:
        return None, str(e)

    return m, None

