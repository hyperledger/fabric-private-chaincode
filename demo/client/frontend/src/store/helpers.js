/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// helper function
module.exports = {
  checkStatus: function(message) {
    return new Promise(function(resolve, reject) {
      if (message.status.rc !== 0) {
        reject(message.status);
      } else {
        resolve(message.response);
      }
    });
  }
};
