/*
* Copyright IBM Corp All Rights Reserved
*
* SPDX-License-Identifier: Apache-2.0
*/
'use strict';

const express = require('express');
const fs = require('fs');
const clockauctionRouter = express.Router();

clockauctionRouter.route('/getDefaultAuction').get (function (request, response) {
    console.log('>>>  Entry:  clockauctionRouter.getDefaultAuction');

    //  read auction.json and return in response
    let result = JSON.parse(fs.readFileSync('auction.json', 'utf8'));
    console.log (result);
    response.json (result);
});


module.exports = clockauctionRouter;
