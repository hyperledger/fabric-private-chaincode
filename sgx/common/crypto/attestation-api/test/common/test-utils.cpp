/*
 * Copyright 2020 Intel Corporation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "test-utils.h"
#include <stdio.h>
#include <sys/stat.h>

#include "error.h"
#include "logging.h"

bool load_file(const char* filename, char* buffer, uint32_t buffer_length, uint32_t* written_bytes)
{
    char* p;
    uint32_t file_size, bytes_read;
    struct stat s;

    FILE* fp = fopen(filename, "r");
    COND2LOGERR(fp == NULL, "can't open file");

    COND2LOGERR(0 > fstat(fileno(fp), &s), "cannot stat file");
    file_size = s.st_size;
    COND2LOGERR(file_size > buffer_length, "buffer too small");

    bytes_read = fread(buffer, 1, file_size, fp);
    COND2LOGERR(bytes_read != file_size, "read bytes don't match file size");
    *written_bytes = bytes_read;

    fclose(fp);
    return true;

err:
    if (fp)
        fclose(fp);

    return false;
}

bool save_file(const char* filename, const char* buffer, uint32_t buffer_length)
{
    FILE* fpo;
    uint32_t bytes;
    fpo = fopen(filename, "w+");
    COND2LOGERR(fpo == NULL, "can't open file");
    bytes = fwrite(buffer, sizeof(uint8_t), buffer_length, fpo);
    COND2LOGERR(bytes != buffer_length, "error bytes written");
    fclose(fpo);
    return true;
err:
    if (fpo)
        fclose(fpo);

    return false;
}
