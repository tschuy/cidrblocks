cidrblocks
==========

Generate subnets within Availability Zones from a CIDR.

CLI Usage
---------

Only the `--cidr` flag is necessary. By default, it splits into four
availability zones and outputs a table:

```
(cidrblocks)(master) → ./cidrblocks --cidr=10.0.0.0/20
VPC Range - 10.0.0.0/20

AZ a (10.0.0.0/22):
	10.0.0.0/23 (Private - 512 addresses)
	10.0.2.0/24 (Public - 256 addresses)
	10.0.3.0/25 (Protected - 128 addresses)
	10.0.3.128/25 (Extra - 128 addresses)

AZ b (10.0.4.0/22):
	10.0.4.0/23 (Private - 512 addresses)
	10.0.6.0/24 (Public - 256 addresses)
	10.0.7.0/25 (Protected - 128 addresses)
	10.0.7.128/25 (Extra - 128 addresses)

AZ c (10.0.8.0/22):
	10.0.8.0/23 (Private - 512 addresses)
	10.0.10.0/24 (Public - 256 addresses)
	10.0.11.0/25 (Protected - 128 addresses)
	10.0.11.128/25 (Extra - 128 addresses)

AZ d (10.0.12.0/22):
	10.0.12.0/23 (Private - 512 addresses)
	10.0.14.0/24 (Public - 256 addresses)
	10.0.15.0/25 (Protected - 128 addresses)
	10.0.15.128/25 (Extra - 128 addresses)
```

The number of availability zones can be specified:

```
(cidrblocks)(master) → ./cidrblocks --cidr=10.0.0.0/20 --azs=3
VPC Range - 10.0.0.0/20

AZ a (10.0.0.0/22):
	10.0.0.0/23 (Private - 512 addresses)
	10.0.2.0/24 (Public - 256 addresses)
	10.0.3.0/25 (Protected - 128 addresses)
	10.0.3.128/25 (Extra - 128 addresses)

AZ b (10.0.4.0/22):
	10.0.4.0/23 (Private - 512 addresses)
	10.0.6.0/24 (Public - 256 addresses)
	10.0.7.0/25 (Protected - 128 addresses)
	10.0.7.128/25 (Extra - 128 addresses)

AZ c (10.0.8.0/22):
	10.0.8.0/23 (Private - 512 addresses)
	10.0.10.0/24 (Public - 256 addresses)
	10.0.11.0/25 (Protected - 128 addresses)
	10.0.11.128/25 (Extra - 128 addresses)

Unused blocks:
	10.0.12.0/22
```

It can also output in four formats:

* `table`
* `cloudformation`
* `terraform`
* `cli` (AWS cli commands) (coming soon)

```
(cidrblocks)(master) → ./cidrblocks --cidr=10.0.0.0/20 --format=cloudformation
{
	"AWSTemplateFormatVersion" : "2010-09-09",
	"Resources" : {
...
```

Web Service
-----------

Alternatively, `cidrblocks` can be run as a web service:

`(cidrblocks)(master) → ./cidrblocks serve --port=8087`

→ `http://localhost:8087/?format=table&cidr=10.0.0.0/8`  
→ `http://localhost:8087/?format=cloudformation&cidr=10.0.0.0/8&azs=2`

Library
-------

`cidrblocks` can also be used as a library within another application with the
`cidrblocks/subnet` subpackage.
