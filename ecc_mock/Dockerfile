FROM hyperledger/fabric-private-chaincode-base-rt:latest

ENV PATH=/opt/ecc:$PATH

WORKDIR /opt/ecc
COPY ecc .

EXPOSE 9999
CMD ["ecc"]