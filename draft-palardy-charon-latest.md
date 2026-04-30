---
title: "CHARON - Cached Host Address Resolution for Outside Networks"
abbrev: "CHARON"
category: info

docname: draft-palardy-charon-latest
submissiontype: IETF  # also: "independent", "editorial", "IAB", or "IRTF"
number:
date:
consensus: true
v: 3
area: AREA
workgroup: WG Working Group
keyword:
 - next generation
 - unicorn
 - sparkling distributed ledger
venue:
  group: WG
  type: Working Group
  mail: WG@example.com
  arch: https://example.com/WG
  github: USER/REPO
  latest: https://example.com/LATEST

author:
 -
    fullname: Andrew Palardy
#    organization: Independent
    email: andrew@apalrd.net

normative:

informative:

...

--- abstract

Cached Host Address Resolution for Outside Networks is a method for combining DNS46 address translation and SIIT packet translation to allow legacy IPv4 hosts to access IPv6-only outside networks. This permits legacy IPv4-only devices to continue to function even as the global internet transitions to providing services over only IPv6.

To fully transition Internet routing from IPv4 to IPv6, translation is currently used to permit hosts which speak only IPv4 or IPv6 to speak to hosts which speak only the other network protoocol. Current translation methods place the burden of IPv4-IPv6 translation on the IPv6-speaking network, thus requiring all inter-network traffic to support IPv4 as a global fallback. It is desirable in the long term to shift this 'default' to requiring all inter-area traffic to support IPv6, with IPv4 support gradually declining. However, it is forseeable that legacy IPv4-only devices will continue to operate for some time and will still require translation by their own network operators.

CHARON provides this method, allowing a network operator with IPv4-only clients to provide access to IPv6-only services by dynamically mapping IPv6-only services resolved via DNS to a local-use IPv4 address on a per address basis.

--- middle

# Conventions and Definitions

{::boilerplate bcp14-tagged}

# Introduction

Currently, all hosts on the Internet implement either IP version 4, IP version 6, or both (a "dual stack" configuration). As IPv4 address resources have been exhausted for some time, IPv4 re-use is becoming a larger concern. While IPv6 is the preferred option, IPv6 does not directly free IPv4 resources - as the Internet operates both versions in parallel, it is currently expected that all services will be accessible via IPv4, and therefore all clients must be able to utilize IPv4, continuing the demand for global IPv4 resources. It is not currently expected that all services will be accessible via IPv6, so networks are able to continue operating using only IPv4.

As operating both protocol families is more difficult than operating only one, it is desirable to implement only the IPv6 protocol. However, networks which operate only the IPv6 protocol still must inter-operate with networks which only operate the IPv4 protocol, requiring protocol translation. Currently, the assumption is that traffic over the internet may always fall back to IPv4, therefore, the burden is on IPv6-only network operators to provide translation services on behalf of IPv4-only networks they are interacting with.

The goal of CHARON is to flip this assumption, and allow IPv4-only networks to provide translation services on their own, freeing IPv6-only networks to remain IPv6-only. 

## Requirement for IP/ICMP Translation

If a source wants to send a packet to dest:
  \      source
   \   v4 | ds| v6
d   \______________
e v4|  B  | B |  C
s ds|  B  | A |  A
t v6|  D  | A |  A

There are four possible scenarios which could be encountered:
A. Both support IPv6 -> IPv6 is preferred and IPv6 used
B. Both support IPv4 -> IPv4 is the global fallback and IPv4 is used
C. Src only supports v6 and dest only supports v4 -> Translation required
D. Src only supports v4 and dest only supports v6 -> Translation required

This translation requirement leaves two options for where to tackle cases C and D using protocol translation - translator provided by the source network or by the destination network.

## Existing Architectures
Existing RFCs define architectures for several translaton scenarios:

For networks with v6-only clients establishing v4 connections, 464XLAT (RFC6877) allows v6-only sources which support a client-side translator (CLAT) to speak to v4-only destinations. This works cleanly across any higher layer protocol and does not depend on DNS or any other side-channel protocol, and is widely deployed by client ISPs. Without a client-side translator, DNS64 (RFC6147) may be used to direct clients to the NAT64 function via DNS. 

For networks with v6-only servers accepting v4 connections, SIIT (RFC7915) to allow the v4-internet to speak to v6-only datacenters, and this is widely deployed as SIIT-DC (RFC7755)

For networks with v4-only servers accepting v6 connections, SIIT(RFC7915) can technically work, but there is no standard for how to compress the IPv6 source address into an IPv4 field to send to the server. TAYGA (github.com/apalrd/tayga) solves this problem by dynamically allocating IPv4 addresses out of a pool, for IPv6 clients. Application-layer gateways (i.e. HTTP proxies) may also be useful here. 

For networks with v4-only clients establishing v6 connections, there is currently no network layer translation method available.

# Translation Method

TODO how it works

# Security Considerations

TODO Security


# IANA Considerations

This document has no IANA actions.


--- back

# Acknowledgments
{:numbered="false"}

TODO acknowledge.