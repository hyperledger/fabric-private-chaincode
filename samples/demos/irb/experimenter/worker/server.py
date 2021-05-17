# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

import sys
import signal
from twisted.web import server, resource, http
from twisted.internet import reactor, defer

import protos.proto_test_pb2 as proto_test
import attestation

class WorkerServer(resource.Resource):
    isLeaf = True

    def render_GET(self, request):
        print('GET REQUEST: %s' % (request.uri).decode("ascii"))
        if request.uri == b'/info':
            request.setHeader(b"content-type", b"text/plain")
            request.setResponseCode(http.OK)
            return "I'm up and running ..".encode("ascii")

        if request.uri == b'/attestation':
            str = attestation.GetAttestation()
            return str.encode("ascii")

        if request.uri == b'/proto-test':
            se = proto_test.ProtoTest()
            se.study_id = "this is a study"
            se.counter = 17
            return se.SerializeToString()

        return ("Unsupported: " + (request.uri).decode("ascii")).encode("ascii")

    def render_POST(self, request):
        print('POST REQUEST: %s' % (request.uri).decode("ascii"))
        if request.uri == b'/proto-test':
            data = request.content.read()
            se = proto_test.ProtoTest()
            se.ParseFromString(data)
            se.counter += 1
            return se.SerializeToString()

        return ("Unsupported: " + (request.uri).decode("ascii")).encode("ascii")        
    

def __shutdown__(*args) :
    print("Shutdown request received")
    reactor.callLater(1, reactor.stop)


def RunWorkerServer():
    
    signal.signal(signal.SIGQUIT, __shutdown__)
    signal.signal(signal.SIGTERM, __shutdown__)

    root = WorkerServer()
    site = server.Site(root)
    reactor.listenTCP(5000, site, interface="0.0.0.0")

    @defer.inlineCallbacks
    def shutdown_twisted():
        print("Stopping Twisted")
        yield reactor.callFromThread(reactor.stop)

    reactor.addSystemEventTrigger('before', 'shutdown', shutdown_twisted)

    try :
        reactor.run()
    except ReactorNotRunning:
        print('exception reactor')
        sys.exit(-1)
    except :
        print('generic exception reactor')
        sys.exit(-1)

    print("Server terminated")
    sys.exit(0)

