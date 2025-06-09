#! /bin/sh
rm dash.log
while true; do
	echo "starting ssh session" >>dash.log
	ssh nas -p 23234 2>>dash.log
	echo "exited ssh session" >>dash.log
	sleep 10
done
