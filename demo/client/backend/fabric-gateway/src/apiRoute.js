/*
* Copyright IBM Corp All Rights Reserved
*
* SPDX-License-Identifier: Apache-2.0
*/
'use strict';

const express = require('express');
const fabricClient = require('./fabricClient');
const ApiRouter = express.Router();

ApiRouter.route('/getRegisteredUsers').get (function (request, response) {
    console.log('>>> GET route api/getRegisteredUsers');
    fabricClient.getRegisteredUsers().then((users) => {
        response.send(users);
    },(error) => {
        response.send(error);
    });
});

module.exports = ApiRouter;
