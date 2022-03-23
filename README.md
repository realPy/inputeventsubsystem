


inputeventsubsystem allow to get low level access to device input interface on Linux platform with go.

It support capabilitities and Abs caps for Device use axis.

inputeventsubsystem Is a part of a big project that need handle event device with high performance, low latency in a high stress condition.

This lib can be usefull for headless project that need handle mouse, keyboard, button etc..

For more information how to handle events, please follow the kernel documentation provide here:

https://www.kernel.org/doc/html/latest/input/input_uapi.html

You can find a Go version of evtest ( https://github.com/freedesktop-unofficial-mirror/evtest/blob/master/evtest.c ) in example directory 