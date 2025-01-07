# Build with Intel SGX HW support

Here we give additional information how to build and run FPC Chaincode with Intel SGX HW support.

## Details - Intel SGX Attestation Support

Note that the simulation mode is for developing purpose only and does
not provide any security guarantees.

As mentioned before, by default the project builds in SGX simulation mode, `SGX_MODE=SIM` as defined in `$FPC_PATH/config.mk` and you can
explicitly opt for building in hardware-mode SGX, `SGX_MODE=HW`. In order to set non-default values for install
location, or for building in hardware-mode SGX, you can create the file `$FPC_PATH/config.override.mk` and override the default
values by defining the corresponding environment variable.

Note that you can always come back here when you want a setup with SGX
hardware-mode later after having tested with simulation mode.

## Register with Intel Attestation Service (IAS)

If you run SGX in __simulation mode only__, you can skip this section.
We currently support EPID-based attestation and  use the Intel's
Attestation Service to perform attestation with chaincode enclaves.

What you need:
* a Service Provider ID (SPID)
* the (primary) api-key associated with your SPID

In order to use Intel's Attestation Service (IAS), you need to register
with Intel. On the [IAS EPID registration page](https://api.portal.trustedservices.intel.com/EPID-attestation)
you can find more details on how to register and obtain your SPID plus corresponding api-key.
We currently support both `linkable` and `unlinkable` signatures for the attestation.

Place your ias api key and your SPID in the `ias` folder as follows:
```bash
echo 'YOUR_API_KEY' > $FPC_PATH/config/ias/api_key.txt
echo 'YOUR_SPID_TYPE' > $FPC_PATH/config/ias/spid_type.txt
echo 'YOUR_SPID' > $FPC_PATH/config/ias/spid.txt
```
where `YOUR_SPID_TYPE` must be `epid-linkable` or `epid-unlinkable`, depending on the type of your subscription.
