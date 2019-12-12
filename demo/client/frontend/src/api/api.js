/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import axios from "axios";

const baseURL = process.env.VUE_APP_API_BASE_URL;

export default axios.create({
  baseURL,
  headers: {
    "Content-Type": "application/json"
  }
});
