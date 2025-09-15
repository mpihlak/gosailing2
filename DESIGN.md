# Elements of the game

## Arena

Two dimensional area on the water. As seen from above, in 2D. Like a map. North
is up. On the arena, the race course is defined and also various static and
dynamic objects can be placed. Also various visualisations may be placed on the
arena and updated regularly.

## Geometry

We use metric units for object sizes and distances. We use nautical units
(knots) for wind and boat speed.

## Wind

Wind blows across the arena. The initial median direction is set to North, so
that it is seen blowing top down on the screen. The wind may shift direction
and strength. The wind may also vary in speed and direction around the race
course. Different objects on the race course can have an impact on the wind.
For example a boat will create a wind shadow.

## Static Objects - Marks

Static objects are placed on the arena. They have a location and size. They do
not change their position or size. They generally do not have impact on wind.

Initially we only consider different "marks". Upwind mark is placed at the top
of the screen. Committee boat mark is placed in the lower right part of the
screen. Pin end mark is in the lower left part of the screen.

## Dynamic objects - Boats

We only consider sailboats here.

Boats have size, location, speed and a heading. Boats also have a tack -
starboard tack if wind is blowing from the starboard side. Port tack if wind is
blowing from the port.

They are generally always moving in the direction of their heading, thus the
location will change. Boat speed depends on the wind speed and angle relative
to the boat heading. We define a polar table that maps wind strength and angle
to boat speed.

Boats can not sail straight into the wind and when pointed in that direction
they will eventually stop. When stopped they will slowly accelerate again when
pointed away from the wind.

Boats cannot pass through other objects such as boats and marks.

## Wind shadows

A sailboat's wind shadow is generally cone or bubble-shaped, extending downwind
from the boat in an area of disturbed air that is smaller and extends further
in lighter wind conditions. Its size and shape also depend on the boat's sail
plan and speed, but the most important factor is that it is cast in the
direction of the boat's apparent wind, not the true wind.

In lighter winds the wind shadow is larger. In stronger winds it is smaller.

Wind shadows of boats will change the wind strength and direction on the arena.
Strength will reduce. The direction will change

## Visualisations

Wind direction and strenght are displayed as equally spaced arrows in grid
structure all over the arena. The direction of the arrow indicates wind
direction. The tail of the arrow indicates strength.

## Race course

Start line. Defined as dotted line drawn between the committee boat and the pin
end mark. Displayed in the lower half of the screen so that there is sufficient
room to define a starting area that can be navigated by the boats.

Upwind mark is placed at the top of the screen in the center.

The race starts at a designated time. After the start time the boats may cross
the starting line by sailing between the committee boat and pin end mark. They
then sail up to the upwind mark, changing tacks as necessary. They round the
upwind mark by leaving it to the port side. Then they sail down back to the
starting line and finish the race by sailing between the committee boat and the
pin end mark.

The boat that finishes first is the winner.

## Start

There is a 5 minute timer that is started before the race start. During that
time the boats position themselves in the best possible position so that they
can cross the starting line at a good position with good speed and on time.

## Implementation

## Technology

The game will be implemented in the Go programming language, using Ebitengine
for graphics. Eventually it will also need to run in a browser on a mobile
device (using WASM). Initial version will run natively.

## Gameplay

Initial version will have the race course and a single boat. Wind is blowing at
a constant speed from North throughout the game.

The boat can be sailed around the race course. Use arrow keys to change
direction in 5 degree increments. Automatically change tacks when passing head
to wind (similarly when going downwind). Calculate boat speed based on the
given table of wind speeds and directions.

When boat is turned head to wind it will gradually stop. It will only
re-accelerate when pointed away from the wind. The acceleration is also
gradual.

In the initial version there is no starting procedure, it is just free sailing.

## Graphics

Simple vector graphics. Light blue background indicates the sea. Little flag
indicates marks. Boats are represented as polygons resembling the shape of
sailboat as seen from above. Sailboats also have a little boom that indicates
the tack they are on.

## Design

The software design of the game shall be modular, with each module being
separately testable. The modules could be go structs. We want structs to
represent the major objects - boat, mark, committe boat, arena, race course,
base wind, wind shadows.

These structs should provide a simple interface that can be later adjusted to
introduce additional behaviors.

## MVP

The wind should be modelled such that the we could ask wind data through an
interface. We would specify the location on the race course and the interface
would return  the wind at that location. The implementation could be later
changed to include wind shifts, shadows, etc. But the interface should not
change much.

Collisions between boats and objects. In the MVP we would handle collisions as
game ending events. The collision detection could be initially very simple (ie.
treat objects as point and radius and boat as simply a point). In the next
iteration we might come up with a more sophisticated approach and also consider
boat on boat collisions.

Controls will be "arcade style" in the MVP. There's no adjustments to sail
trim, etc.

The boat should initially move at a speed that would be equivalent of 6 kts
boatspeed in the real world.

Interaction and game state. There would be an overall GameState central
structure that keeps track of boat locations, marks, general wind as well as
specific wind conditions. The GameState supports the methods require by the
Ebitengine, so it would have Update, Draw and other required methods.

Directory layout. Use cmd for any binaries and pkg for projects packages.
