Let's go back to implementing a coordinate system. As before create a struct
that has methods for converting between world and screen coordinates. Do not
migrate anything over to using world coordinates just yet. 

The world coordinates are in a system where 1 unit is 1 meter.  It should use a
scaling factor such that 180 meters (the intended length of the starting line)
fits on the screen. When converting between the pixel and world coordinates we
need to consider the aspect ratio of the screen layout.

The coordinate system should have a concept of camera or observer. 
