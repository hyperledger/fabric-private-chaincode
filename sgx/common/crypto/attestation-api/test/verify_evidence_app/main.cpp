/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "test.h"

int main()
{
    bool b = test();
    if (b)
    {
        // success
        return 0;
    }
    // error
    return -1;
}
