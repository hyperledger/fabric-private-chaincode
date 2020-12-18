/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#define COND2ERR(b)                                            \
    do                                                         \
    {                                                          \
        if (b)                                                 \
        {                                                      \
            LOG_DEBUG("error at %s:%d\n", __FILE__, __LINE__); \
            goto err;                                          \
        }                                                      \
    } while (0)

#define COND2LOGERR(b, msg)                                             \
    do                                                                  \
    {                                                                   \
        if (b)                                                          \
        {                                                               \
            LOG_ERROR("error at %s:%d: %s\n", __FILE__, __LINE__, msg); \
            goto err;                                                   \
        }                                                               \
    } while (0)

#define CATCH(b, expr) \
    do                 \
    {                  \
        try            \
        {              \
            expr;      \
            b = true;  \
        }              \
        catch (...)    \
        {              \
            b = false; \
        }              \
    } while (0);
