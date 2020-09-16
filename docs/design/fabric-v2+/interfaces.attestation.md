# Attestation Interface

This document defines the interface for producing and verifying an attestation.


## Attestation

This interface is only available inside an enclave.

```
bool init_attestation(
    uint8_t* params,
    uint32_t params_length);

```
The `init_attestation` accepts as input a binary array of (possibly encoded) parameters. It returns `false` on error.
```
bool get_attestation(
    uint8_t* statement,
    uint32_t statement_length,
    uint8_t* attestation,
    uint32_t attestation_max_length
    uint32_t* attestation_length);
```
The `get_attestation` accepts as input a statement and the buffer where the output attestation will be placed. It returns `false` on error.
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
    [in, size=sig_rl_len] uint8_t *sig_rl, uint32_t sig_rl_len,
    uint32_t sign_type,
    [in, size=report_len] uint8_t *report, uint32_t report_len,
    [out, size=max_quote_len] uint8_t *quote, uint32_t max_quote_len,
    [out] uint32_t *actual_quote_len);
```

3. The output of `get_attestation` is IAS-verifiable, and not publicly verifiable.
So it cannot be provided directly to `verify_evidence` for verification.
Rather, it is delegated to different entity the task of sending the IAS-verifiable attestation to IAS.
This step involves contacting and authenticating with a web service using the ISV's credentials.
IAS will then convert the attestation in a publicly-verifiable report, which can be provided to `verify_evidence`.


## Attestation-to-Evidence Conversion

Depending on the type of attestation that was originally requested,
the attestation may not be immediately publicly-verifiable.
For this reason, it is required an attestation-to-evidence conversion, namely:
```
evidence <- AttestationToEvidence(attestation)
```

### Details related to EPID-based SGX attestations

This step is mainly needed to support EPID-based attestations.
Such attestations are in fact IAS-verifiable quotes which must be converted by contacting IAS.
The owner of the SPID (used in the quote) and the API key (used for authenticating with IAS) can directly send the quote to IAS, which will return a publicly verifiable report (if the quote verification is successful).


## Evidence Verification

This interface is available both inside and outside of an enclave.

```
bool verify_evidence(
    uint8_t* evidence,
    uint32_t evidence_length,
    uint8_t expected_statement,
    uint32_t expected_statement_length,
    uint8_t* expected_code_id,
    uint32_t expected_code_id_length);
```
The `verify_evidence` accepts as input the (publicly-verifiable) evidence to be verified,
the expected statement computed by the caller (which will have to match the attestation statement),
and the expected identity of the code computed by the caller (which will have to match the code identity included in the attestation).
It returns `false` on error.

### Details related to EPID-based SGX attestations

The function caller supplies:
1. the `evidence` parameter containing the publicly-verifiable report issued by IAS.
This is different than the output of the `get_attestation`, which is the IAS-verifiable quote.
2. the expected statement, which typically is the concatenation of some public keys.
3. the expected code identity which, in this case, is a hash that must match MRENCLAVE field in the attestation.
