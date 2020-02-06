/*
* Copyright IBM Corp All Rights Reserved
*
* SPDX-License-Identifier: Apache-2.0
*/

'use strict';

const fs = require('fs');
const fabricClient = require('./fabricClient');
const ccRoute = require('./ccRoute');
const apiRoute = require('./apiRoute');
const clockauctionRoute = require('./clockauctionRoute');

let express = require('express');
let app = express();
app.use(express.json());
app.use(function (req, res, next) {
    res.header('Access-Control-Allow-Origin', '*');
    res.header('Access-Control-Allow-Headers', 'Access-Control-Request-Method, Access-Control-Request-Headers, Origin, X-Requested-With, Content-Type, Accept, x-user');
    res.header('Access-Control-Allow-Methods', 'GET, POST');
    next();
});

app.get('/',(function(req,res){
    res.send('Welcome to Spectrum Auction by Elbonia Communication Commission - Enabled by FPC');
}));

//  different routes defined
app.use('/api/cc', ccRoute);
app.use('/api', apiRoute);
app.use('/api/clock_auction', clockauctionRoute);

async function main() {
    let port = JSON.parse(fs.readFileSync('config.json', 'utf8'))['backend_port'];

    try {
        await fabricClient.connectToNetwork ();
    } catch (error) {
        console.log ('Error in connecting to Fabric network. ', error);
        process.exit (1);
    }

    if (port !== undefined) {
        console.log ('Listening on localhost:' + port + '...');
        app.listen(port,function() {});
    }
    else {
        console.log ('Port not defined; Aborting backend server...');
        process.exit (1);
    }

}


main();
