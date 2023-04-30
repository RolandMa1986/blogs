#!/bin/bash

gcc -shared -fPIC -I/usr/include/python3.8 spammodule.c  -o ./spam.so

## demo load
python3 spam_tutorial.py 
 