#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
#include <string.h>
#include <arpa/inet.h>
#include <pcap.h>
#include "tcpip.h"

#define MYSQLX_PORT 33060
#define MYSQLX_CLIENT 0
#define MYSQLX_SERVER 1

struct streams {
    FILE *f[2];
    unsigned short port_dst;
};

pcap_t* open_device(char const *name)
{
    char errbuf[PCAP_ERRBUF_SIZE];
    pcap_t *handle = pcap_open_live(name, BUFSIZ, 0, 1000, errbuf);
    if (!handle) {
        fprintf(stderr, "pcap_open_live: %s\n", errbuf);
        exit(EXIT_FAILURE);
    }
    return handle;
}

void deal_packet(u_char *arg, const struct pcap_pkthdr *pkthdr, const u_char *packet)
{
    struct sniff_ip *ip = (struct sniff_ip *)(packet + ETHER_SIZE);
    char ip_src[32], ip_dst[32];
    strcpy(ip_src, inet_ntoa(ip->ip_src));
    strcpy(ip_dst, inet_ntoa(ip->ip_dst));
    int ip_size = IP_HEADER_LENGTH(ip);

    struct sniff_tcp *tcp = (struct sniff_tcp *)(packet + ETHER_SIZE + ip_size);
    unsigned short port_src = htons(tcp->tcp_sport);
    unsigned short port_dst = htons(tcp->tcp_dport);
    int tcp_size = TCP_OFF(tcp);

    struct streams *ss = (struct streams *)arg;

    char name_buf[64];
    bzero(name_buf, 64);
    if ((!ss->f[MYSQLX_CLIENT] && port_dst == ss->port_dst) || (!ss->f[MYSQLX_SERVER] && port_src == ss->port_dst)) {
        sprintf(name_buf, "%s:%d_to_%s:%d.tcpdump", ip_src, port_src, ip_dst, port_dst);
    }

    FILE *fp;
    if (port_dst == MYSQLX_PORT) {
        if (!ss->f[MYSQLX_CLIENT] ) {
            fp = fopen(name_buf, "wb");
            assert(fp);
            ss->f[MYSQLX_CLIENT] = fp;
        } else {
            fp = ss->f[MYSQLX_CLIENT];
        }
    } else {
        if (!ss->f[MYSQLX_SERVER] ) {
            fp = fopen(name_buf, "wb");
            assert(fp);
            ss->f[MYSQLX_SERVER] = fp;
        } else {
            fp = ss->f[MYSQLX_SERVER];
        }
    }
    u_char const *payload = packet + ETHER_SIZE + ip_size + tcp_size;
    int payload_len = pkthdr->len - (payload - packet);
    fwrite(payload, 1, payload_len, fp);
    fflush(fp);
}

int main(int argc, char* argv[])
{
    if (argc != 3) {
        fprintf(stderr, "usage: dumper interface port");
        exit(EXIT_SUCCESS);
    }

    char *device = argv[1];
    pcap_t *handle = open_device(device);
    printf("open %s ok\n", device);

    char bpf_buffer[256];
    sprintf(bpf_buffer, "port %s", argv[2]);

    struct bpf_program filter;
    if (pcap_compile(handle, &filter, bpf_buffer, 1, 0) < 0) {
        fprintf(stderr, "pcap_compile: %s\n", pcap_geterr(handle));
        exit(EXIT_FAILURE);
    }
    if (pcap_setfilter(handle, &filter) < 0) {
        fprintf(stderr, "pcap_setfilter: %s\n", pcap_geterr(handle));
        exit(EXIT_FAILURE);
    }

    struct streams ss;
    bzero(&ss, sizeof(struct streams));
    ss.port_dst = atoi(argv[2]);
    pcap_loop(handle, -1, deal_packet, (u_char *)&ss);
    pcap_close(handle);
    return 0;
}
