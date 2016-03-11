#ifndef IP_TABLES_H
#define IP_TABLES_H
#include <stdint.h>

#define IP_ADDR_MIN_LENGTH 8
#define MAX_PORT_NUM 0xFFFF
#define INPUT_CHAIN "INPUT"
#define RULE_NAME_SIZE 64

typedef struct rule_entry_s {
    char *Name; 
    char *PhysicalPort; 
    char *Action; 
    char *IpAddr; 
    char *Protocol; 
    uint16_t  Port;
    int  PrefixLength;
}rule_entry_t;

typedef struct ipt_config_s {
    char   name[RULE_NAME_SIZE];
    struct ipt_entry *entry;
}ipt_config_t;

// ADD RULE
int add_iptable_tcp_rule(rule_entry_t *config, ipt_config_t *rc);
int add_iptable_udp_rule(rule_entry_t *config, ipt_config_t *rc);

// DELETE RULE
int del_iptable_rule(ipt_config_t *config);

#endif
