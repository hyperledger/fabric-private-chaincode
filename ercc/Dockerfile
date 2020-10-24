FROM hyperledger/fabric-private-chaincode-base-rt:latest

ENV PATH=/opt/ercc:$PATH

WORKDIR /opt/ercc
COPY ercc .

EXPOSE 9999
CMD ["ercc"]