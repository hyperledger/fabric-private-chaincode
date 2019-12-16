/*
* Copyright IBM Corp All Rights Reserved
*
* SPDX-License-Identifier: Apache-2.0
*/

// ccRouter.js
'use strict';

const express = require('express');
const fabricClient = require('./fabricClient');
const ChaincodeRouter = express.Router();

function getTransactionArguments (request) {
    let tx = request.body.tx;
    let args = request.body.args;

    let userName = ( request.headers['x-user']);
    userName = userName.trim();

    //  Insert transaction name, userName at index 1,0 in args array
    args.unshift(tx);
    args.unshift(userName);
    return args;
}

ChaincodeRouter.route('/invoke').post(function (request, response) {
    console.log('>>> POST route api/cc/invoke');
    let args = getTransactionArguments (request);

    fabricClient.invoke.apply ('unused', args)
        .then(function (result)  {
            response.json (result);
        })
        .catch(error => {
            response.status(error.status.rc).send(error);
        }
        );
});

ChaincodeRouter.route('/query').post(function (request, response) {
    console.log('>>> POST route api/cc/query');
    let args = getTransactionArguments (request);

    fabricClient.query.apply ('unused', args)
        .then(function (result)  {
            response.json (result);
        })
        .catch(error => {
        //  if enrollUser failed then send 401 Unauthorized
            response.status(error.status.rc).send(error);
        });

});

module.exports = ChaincodeRouter;
