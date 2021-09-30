# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

import sys
import signal
from twisted.web import server, resource, http
from twisted.internet import reactor, defer

import attestation
import server_api
import diagnose

class WorkerServer(resource.Resource):
    isLeaf = True

    def render_GET(self, request):
        print('GET REQUEST: %s' % (request.uri).decode("ascii"))
        if request.uri == b'/info':
            request.setHeader(b"content-type", b"text/plain")
            request.setResponseCode(http.OK)
            return "I'm up and running ..".encode("ascii")

        if request.uri == b'/attestation':
            return server_api.HandleAttestation()

        if request.uri == b'/shutdown':
            __shutdown__()
            return request.uri

        if request.uri == b'/test-diagnose':
            data = b'37.7, 0, 0, 1, 1, 0'
            try:
                result = diagnose.diagnose(data)
            except Exception as e:
                return str(e).encode("ascii")

            if result[0] == "":
                return result[1].encode("ascii")
            else:
                return result[0].encode("ascii")


        return ("Unsupported: " + (request.uri).decode("ascii")).encode("ascii")

    def render_POST(self, request):
        print('POST REQUEST: %s' % (request.uri).decode("ascii"))

        if request.uri == b'/execute-evaluationpack':
            try:
                #data should contain the encrypted evaluation pack protobuf
                data = request.content.read()
                result = server_api.HandleExecuteEvaluationPack(data)
                return result.encode("ascii")

            except Exception as e:
                request.setResponseCode(400)
                return ("Error: " + str(e)).encode("ascii")

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

