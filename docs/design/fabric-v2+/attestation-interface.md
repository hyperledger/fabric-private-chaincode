# Attestation Interface

This document defines the interface for producing and verifying an attestation.


## Producing an attestation

This interface is only available inside an enclave.

```
int init_attestation(
    uint8_t* params,
    uint32_t params_length);

```
The `init_attestation` accepts as input a binary array of (possibly encoded) parameters. It returns `0` on error.
```
int get_attestation(
    uint8_t* statement,
    uint32_t statement_length,
    uint8_t* attestation,
    uint32_t attestation_length);
```
The `get_attestation` accepts as input a statement and the buffer where the output attestation will be placed. It returns `0` on error.
The statement is the object of the attestation. Its content is entirely up to the caller.

### Details related to EPID-based SGX attestations

FPC can use EPID-based SGX attestations.
In this case, the `statement` will be hashed and included in the report data field of the attestation.

EPID attestations have some peculiarities.
1. They require to be initialized with some external parameters (SPID, signature revocation list). These parameters are provided through `init_attestation`.

2. They are computed indirectly through a different enclave, called Quoting Enclave. For this reason, the implementation of `get_attestation` uses the following edge functions to retrieve the IAS-verifiable attestation.
```
void ocall_init_quote(
    [out, size=target_len] uint8_t *target, uint32_t target_len,
    [out, size=egid_len] uint8_t *egid, uint32_t egid_len);

void ocall_get_quote(
    [in, size=spid_len] uint8_t *spid, uint32_t spid_len,
    [in, size=report_len] uint8_t *report, uint32_t report_len,
    [out, size=max_quote_len] uint8_t *quote, uint32_t max_quote_len,
    [out] uint32_t *actual_quote_len);
```

3. It must to be sent to IAS, which provides a publicly verifiable report. This step thus involves contacting and authenticating with a web service using the ISV's credentials. For this reason, its completion is delegated to a different entity.


## Verifying an attestation

This interface is available both inside and outside of an enclave.

```
int verify_attestation(
    uint8_t* attestation,
    uint32_t attestation_length,
    uint8_t expected_statement,
    uint32_t expected_statement_length,
    uint8_t* expected_code_id,
    uint32_t expected_code_id_length);
```
The `verify_attestation` accepts as input the attestation to be verified,
the expected statement computed by the caller (which will have to match the attestation statement),
and the expected identity of the code computed by the caller (which will have to match the code identity included in the attestation).
It returns `0` on error.

### Details related to EPID-based SGX attestations

The `attestation` parameter contains the publicly verifiable report issued by IAS.
The function caller computes the expected statement, usually as the hash of some data.
Such caller also provides the expected code identity which, in this case, is a hash that must match MRENCLAVE field in the attestation.
