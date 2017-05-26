#/bin/bash


start() {
    echo 'build sys'

    `which go` build

    echo 'start sys'

    nohup ./kvmtop  2>&1 > kvmtop.log &
}

stop(){
    for pid in `ps x | grep kvmtop|grep -v grep | awk '{print $1}'`;do
        echo "SHUTDOWN process $pid"
        kill -9 $pid
    done
}

case $1 in
stop)
    stop 
    ;;
start)
    start 
    ;;
restart)
    stop 
    start 
    ;;
*)
   echo "kvmtop start|stop|restart"
   ;;
esac