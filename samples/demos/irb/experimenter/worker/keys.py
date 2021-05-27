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

