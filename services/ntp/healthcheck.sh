#!/bin/sh
# Healthy when chrony has selected a source and is synchronised.
chronyc -n tracking | grep -q "Leap status     : Normal"
