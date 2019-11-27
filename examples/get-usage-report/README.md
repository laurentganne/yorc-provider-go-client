# Getting infrastructure usage reports

This example shows how the yorcprovider go client can be used to query and display
usages reports on a given location.

## Prerequisites

A premium plugin **alien4cloud-yorc-collector-plugin** has to be uploaded in [Alien4Cloud](http://alien4cloud.github.io/#/documentation/2.1.0/user_guide/plugin_management.html),
contact [Yorc team](https://gitter.im/ystia/yorc?source=orgpage) for details.

A plugin able to report infrastructure usages has to be uploaded in [Yorc](https://yorc.readthedocs.io/en/latest/plugins.html).
For example the [HEAppE plugin](https://github.com/laurentganne/yorc-heappe-plugin)
has the ability to report infrastructure usages.

## Running this example

Build this example:

```bash
cd examples/get-usage-report
go build -o collect.test
```

Now, run this example providing in arguments:
* the Alien4Cloud URL
* credentials of a user who needs to have the administrator role
* the name of the orchestrator managing the location for which we will get a report
* the type of infrastructure of the location for which we will get a report
* the name of the location for which we will get a report
* optionally query parameters, depending on the plugin implementation

For example, for a location managed by the [HEAppE plugin](https://github.com/laurentganne/yorc-heappe-plugin),
this command will report the current usage of cluster nodes:

```bash
./collect.test -url https://1.2.3.4:8088 \
               -user myuser \
               -password mypasswd \
               -orchestrator Yorc \
               -type heappe \
               -location myHeappeLocation

```

If this command is called with additional query parameters, it will reports Job resources usages
for a given period (with **start** and **end** query parameters), and for a given user
(with **user** query parameter, if not specified the user specified in Yorc configuration for this location will be used).
For example, to get a Job resources usage report from January 1st 2019, to November 11 2019: 

```bash
./collect.test -url https://1.2.3.4:8088 \
               -user myuser \
               -password mypasswd \
               -orchestrator Yorc \
               -type heappe \
               -location myHeappeLocation \
               -query "start=2019-09-01" \
               -query "end=2019-11-27"
```
