# Winch box revision — base motors and roof-mounted electronics

**Implemented geometry is now available in [hinged-roof-v4](../hinged-roof-v4/README.md).**
It uses a complete hinged upper hood so the carrier clears the opening. The v3
CAD below remains the historical removable-lid version.

## Confirmed arrangement

- Both motors and gearboxes mount on a rigid metal base/support structure.
- Both drivers, the 48-to-24 V converter and two brake relays mount on a metal
  carrier attached beneath a hinged roof, above the motors when closed.
- Opening the roof lifts the electronics with it and exposes the motor bay.
- The separate power/Jetson container and 2 m detachable hybrid cable remain
  as specified in ELECTRICAL.md.

## Mechanical design direction

Use a rear-hinged roof, opening away from the drums, with metal hinges and a
positive hold-open stay sized for the complete roof/carrier mass. Support the
driver heatsinks and DIN rail on metal brackets attached to the roof carrier;
do not hang the equipment from unsupported printed sheet or printed hinges.
Keep devices in their required operating orientation with the roof closed.

The continuous TPU perimeter gasket needs uniform compression from closing
fasteners or compression latches around the perimeter. Hinges alone do not
provide this. Place hinge fixings outside the seal where possible; otherwise
design sealed penetrations and suitable backing plates.

The current 186 mm top opening, hanging electronics and overlapping lid skirt
must be redesigned together. Check the entire carrier, equipment and connector
sweep from closed to the stay's opening angle against the rim, body and motors.
Do not assume the existing equipment positions can rotate through the rim.
Consider an offset/lift-clear hinge or revised skirt geometry to avoid rear-edge
interference and dragging the gasket. Establish the pivot only after checking
this sweep. Reserve rear space for the open roof.

Mount the hybrid bulkhead connector on the fixed body, so its external cable
does not move when opening the lid. Bring the incoming power/data and motor,
encoder and brake wiring through restrained flexible service loops near the
hinge. Separate motor wiring from signal pairs, respect cable bend radii and
provide strain relief on both moving and fixed ends. Keep loops away from the
gasket, hinge pinch points and the rotating shafts. Include a flexible bonding
strap where required; do not rely on hinge bearings for electrical bonding.

The roof-mounted drivers need a heat path through the moving metal carrier to
an appropriately sized external heat sink/plate. The previously reserved fixed
rear thermal interface is not automatically connected to that moving carrier.
Include sink mass in the hinge/stay calculation and preserve rain sealing at
the thermal interface. Motor losses still require their own heat path.

## Artifact status

This document records the new mounting requirement. The existing STLs, GLB and
layout drawing still show the earlier removable-lid concept and **do not yet
implement the hinged roof or its mounting carrier**. Revise the hinge geometry,
opening, carrier and wiring envelopes and validate their motion before printing
the winch lid/riser for this arrangement. Component mounting patterns, hinge/stay
hardware and thermal hardware remain to be selected against actual parts.
