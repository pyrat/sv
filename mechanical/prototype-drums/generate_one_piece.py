"""Generate the one-piece PETG demonstration drum; dimensions are millimetres.

Uses the shared shaft clamp and STL checks from generate.py. Run with the same
requirements.txt environment. Print two copies of drum-one-piece.stl.
"""
import manifold3d as m

from generate import CLAMP_HEIGHTS, SEGMENTS, box, clamp, cylinder, revolve, save


def access_slot():
    # A 10.5 mm wide vertical capsule admits M4 heads AND 9 mm OD washers.
    # Its endpoints align with the two bolt axes. Keep the cutter short of
    # the clamp lugs, which start at Y=-12.
    low, high = CLAMP_HEIGHTS
    slot = box((10.5, 40, high - low), (18 - 5.25, -60, low))
    for z in CLAMP_HEIGHTS:
        slot += cylinder(5.25, 40).rotate((90, 0, 0)).translate((18, -20, z))
        entry = m.Manifold.cylinder(11, 5.25, 6, SEGMENTS)
        slot += entry.rotate((90, 0, 0)).translate((18, -31, z))
    # Join the flared ends with a tapered rectangular passage.
    section = m.CrossSection([[(12, -42), (24, -42),
                              (23.25, -31), (12.75, -31)]])
    slot += section.extrude(high - low).translate((0, 0, low))
    return slot


def main():
    # Lower flange on bed. Upper flange grows radially by 15 mm over 15 mm
    # height: a 45-degree overhang, with no roof over the 64 mm centre opening.
    profile = [(0, 0), (54.2, 0), (55, 0.8), (55, 5.2),
               (54.2, 6), (41, 6), (40, 7), (40, 31),
               (55, 46), (55, 49.2), (54.2, 50), (32.8, 50),
               (32, 49.2), (32, 6), (0, 6)]
    body = clamp(revolve(profile)) - access_slot()
    for z in CLAMP_HEIGHTS:
        # Clear path for the whole 9 mm washer/head envelope to the lug face,
        # rather than merely checking the hex key fits through the barrel.
        hardware = cylinder(4.5, 48).rotate((90, 0, 0)).translate((18, -12, z))
        assert (body ^ hardware).volume() < 0.001, 'Bolt/washer insertion blocked'
        shaft = cylinder(2, 30).rotate((90, 0, 0)).translate((18, 18, z))
        assert (body ^ shaft).volume() < 0.001, 'M4 bolt shaft obstructed'
    # The flange must leave the centre open above the clamp for nut insertion.
    opening = cylinder(31.9, 21, 30.1)
    assert (body ^ opening).volume() < 0.001, 'Central hardware access obstructed'
    save('drum-one-piece.stl', body)
    save('shaft-fit-test.stl', clamp(cylinder(26, 6)))


if __name__ == '__main__':
    main()
