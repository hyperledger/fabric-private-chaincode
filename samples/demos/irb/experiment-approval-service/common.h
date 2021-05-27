/*
 * Copyright 2019 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

//#define DO_DEBUG true

#include <list>
#include <numeric>
#include <queue>
#include <set>
#include <string>
#include <vector>
#include "json/parson.h"
#include "logging.h"
#include "shim.h"

// needed for types and Base64 conversion primitives
// TODO: see if this can be moved in shim
#include "../../../../common/crypto/pdo/common/types.h"

// needed for crypto primitives
// TODO: see if this can be moved in shim
#include "../../../../common/crypto/pdo/common/crypto/crypto.h"
