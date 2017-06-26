#ifndef PROTOCOL_H_
#define PROTOCOL_H_
#include <stdlib.h>
#include <netinet/in.h>

#define ETHER_SIZE 14
#define ETHER_ADDR_LEN  6

struct sniff_ethernet {
    u_char ether_dhost[ETHER_ADDR_LEN]; /* Destination host address */
    u_char ether_shost[ETHER_ADDR_LEN]; /* Source host address */
    u_short ether_type;                 /* IP? ARP? RARP? etc */
};

struct sniff_ip {
    u_char ip_vhl;          /* version << 4 | header length >> 2 */
    u_char ip_tos;          /* type of service */
    u_short ip_len;         /* total length */
    u_short ip_id;          /* identification */
    u_short ip_off;         /* fragment offset field */
    u_char ip_ttl;          /* time to live */
    u_char ip_p;            /* protocol */
    u_short ip_sum;         /* checksum */
    struct in_addr ip_src;
    struct in_addr ip_dst;  /* source and dest address */
};

#define IP_RESERVED_FLAG 0x8000 /* reserved fragment flag */
#define IP_DONT_FLAG 0x4000     /* dont fragment flag */
#define IP_MORE_FLAG 0x2000     /* more fragments flag */
#define IP_OFFMASK 0x1fff       /* mask for fragmenting bits */

#define IP_VERSION(ip)          (((ip)->ip_vhl) >> 4)
#define IP_HEADER_LENGTH(ip)    (((ip)->ip_vhl) & 0x0f)

/* TCP header */
struct sniff_tcp {
    u_short tcp_sport;   /* source port */
    u_short tcp_dport;   /* destination port */
    u_int32_t tcp_seq;   /* sequence number */
    u_int32_t tcp_ack;   /* acknowledgement number */

    u_char tcp_offx2;    /* data offset, rsvd */
    u_char tcp_flags;
    u_short tcp_win;     /* window */
    u_short tcp_sum;     /* checksum */
    u_short tcp_urp;     /* urgent pointer */
};

#define TCP_OFF(th)  (((th)->tcp_offx2 & 0xf0) >> 4 << 2)
#define TCP_FIN 0x01
#define TCP_SYN 0x02
#define TCP_RST 0x04
#define TCP_PUSH 0x08
#define TCP_ACK 0x10
#define TCP_URG 0x20
#define TCP_ECE 0x40
#define TCP_CWR 0x80
#define TCP_FLAGS (TCP_FIN|TCP_SYN|TCP_RST|TCP_ACK|TCP_URG|TCP_ECE|TCP_CWR)

#endif
