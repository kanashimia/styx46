#enable proxy-nd
echo "1" > /proc/sys/net/ipv6/conf/eth0/proxy_ndp
echo "2" > /proc/sys/net/ipv6/conf/eth0/accept_ra
echo "1" > /proc/sys/net/ipv6/conf/all/forwarding
echo "1" > /proc/sys/net/ipv4/conf/all/forwarding

#add 16 proxy entries
ip -6 neigh add proxy 2001:99a:390:2800::460 dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::461 dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::462 dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::463 dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::464 dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::465 dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::466 dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::467 dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::468 dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::469 dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::46a dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::46b dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::46c dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::46d dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::46e dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::46f dev eth0
ip -6 neigh add proxy 2001:99a:390:2800::470 dev eth0