#!/bin/sh
cd /home/pi/e-paper-inflaxdb/main
export V2_ROTATE_180=false
export INIT_CLEAR_FLAG=false
export V2_IMG_MIRROR=true
export V2_FLAG=true
export TRASH_FLAG=true
./epaperifdb
