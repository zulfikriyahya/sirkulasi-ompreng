#!/bin/sh
udevadm control --reload-rules 2>/dev/null || true
udevadm trigger 2>/dev/null || true
