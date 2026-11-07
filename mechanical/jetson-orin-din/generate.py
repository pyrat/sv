"""Edge-on Orin Nano DIN bracket, millimetres. Prototype: physical fit untested.

Print coordinates: PCB in XY, fan side +Z, GPIO +Y, USB +X.
DIN clip on the bottom 79 mm edge (-Y); rail length parallel to Z.
"""
from pathlib import Path
import json
import manifold3d as m
import numpy as np
import trimesh

OUT = Path(__file__).parent
GAP = 0.5
STOP_D = 3.0
STOP_PILOT_D = 2.5
STOP_X = -5.0
STOP_Y = 2.0
SLIDER_Z = 2.5
PILOT_D = 2.15
PCB_Z = 14.0
HOLES = [(4, 4), (62, 4), (4, 96), (62, 96)]


def box(size, at=(0, 0, 0)):
    return m.Manifold.cube(size).translate(at)


def cyl(radius, height, at):
    return m.Manifold.cylinder(height, radius, radius, 64).translate(at)


def mesh(s):
    data = s.to_mesh()
    return trimesh.Trimesh(np.asarray(data.vert_properties)[:, :3],
                            np.asarray(data.tri_verts), process=True)


def short_edge(s):
    """Map clip-local axes to the short edge, centred at PCB X=39.5."""
    return s.rotate((0, 0, 90)).translate((89.5, -2, 0))


def make_parts():
    # Open ladder frame: 10 mm component clearance above its 4 mm ribs.
    body = box((5.6, 100, 4), (1.4, 0, 0))
    body += box((8, 100, 4), (58, 0, 0))
    body += box((75.1, 8, 4), (1.4, 0, 0))
    body += box((64.6, 8, 4), (1.4, 92, 0))
    for x, y in HOLES:
        body += cyl(3, PCB_Z, (x, y, 0))
        body -= cyl(PILOT_D / 2, 9, (x, y, PCB_Z - 8))

    clip, slider = make_clip()
    # Connect the relocated spine to the bottom frame rib. The 2 mm setback
    # keeps the moving latch clear of the bottom-right standoff and frame.
    body += box((35.5, 4.5, 4), (21.5, -0.5, 0))
    body += short_edge(clip)
    return body, short_edge(slider)


def make_clip():
    # Clip-local geometry; short_edge() rotates the whole mechanism together.

    # Rail length is Z; flange front face X=-6, edges Y=32.5 and 67.5.
    # Deep spine spreads the overturning load along 24 mm of rail.
    body = box((7.8, 35.5, 4), (-6, 32.5, 0))
    body += box((5, 35.5, 24), (-6, 32.5, 0))
    body += box((9, 3, 24), (-10, 68, 0))
    body += box((2, 5.5, 24), (-10, 65.5, 0))

    # Four-sided tunnel: 0.4 mm side gaps and 0.5 mm above/below the slider.
    # The solid, full-width slider and hook enter from local -Y.
    body += box((2.6, 31, 11), (1.4, -2, 0))
    body += box((2.6, 24.8, 11), (-13, -2, 0))
    body += box((17, 24.8, 2), (-13, -2, 0))
    body += box((17, 24.8, 3), (-13, -2, 8))
    # Rear shoulder joins the tunnel to the spine and stops inward travel.
    body += box((2.6, 6, 4), (1.4, 28, 0))
    body += box((10, 3, 8), (-6, 32, 0))
    # M3 x 8 screw crosses the entry BEHIND the solid slider tail.
    # It never passes through the slider; remove it to withdraw the slider.
    body -= cyl(STOP_PILOT_D / 2, 4, (STOP_X, STOP_Y, 7.5))
    # Lock pocket beside the accessible flexure tongue; floor stays intact.
    body -= box((1.2, 4.8, 6), (1.4, 7.6, 2))
    # Narrow top access window for pressing the tongue with a flat screwdriver.
    # The roof still captures the full-width rigid spine on the opposite side.
    body -= box((2.8, 12, 4), (-0.2, 1, 8))

    # Solid tail supplies the external screw's withdrawal-stop face.
    slider = box((11, 21.8, 5), (-10, 10, 0))
    # Integral 1.2 mm-wide cantilever, 18 mm free length; press toward -X.
    # 1.4 mm flexure relief allows the tooth's required ~1 mm retraction.
    slider -= box((1.4, 18, 7), (-1.6, 10, -1))
    slider += box((2.4, 4, 5), (-0.2, 8, 0))
    # Moving lower hook; its outside elbow is clear of the guide at full travel.
    slider += box((5.2, 2.3, 5), (-10, 29.5, 0))
    slider += box((2, 5, 5), (-10, 29.5, 0))
    return body, slider.translate((0, 0, SLIDER_Z))


def rail():
    # Nominal 35 x 7.5 mm rail with 1 mm stock, for interference checks.
    r = box((1, 25, 44), (-13.5, 37.5, -10))
    for y in (37.5, 61.5):
        r += box((7.5, 1, 44), (-13.5, y, -10))
    for y in (32.5, 62.5):
        r += box((1, 5, 44), (-7, y, -10))
    return short_edge(r)


def main():
    body, slider = make_parts()
    clip, _ = make_clip()
    clip = short_edge(clip)
    parts = [('body', body), ('slider', slider), ('clip-test', clip)]
    for name, part in parts:
        assert part.status() == m.Error.NoError, name
        assert len(part.decompose()) == 1, name
        assert mesh(part).is_volume, name
    assert (body ^ slider).volume() < 1e-6
    assert (body ^ rail()).volume() < 1e-6
    assert (slider ^ rail()).volume() < 1e-6
    assert (clip ^ slider).volume() < 1e-6
    assert (clip ^ rail()).volume() < 1e-6
    pcb = box((79, 100, 1.57), (0, 0, PCB_Z))
    assert (body ^ pcb).volume() < 1e-6

    # Test rigid guide/hook travel; locking tongue must be pressed separately.
    rigid = slider - short_edge(box((4, 20, 9), (-1.6, 8, 0)))
    stop = short_edge(cyl(STOP_D / 2, 5, (STOP_X, STOP_Y, 3)))
    # With screw absent and tongue depressed, slide in from the right.
    for insertion in np.linspace(0, 38, 381):
        assert (body ^ rigid.translate((float(insertion), 0, 0))).volume() < 1e-6, insertion
        assert (clip ^ rigid.translate((float(insertion), 0, 0))).volume() < 1e-6, insertion
    for travel in np.linspace(0, 6.5, 66):
        moved = rigid.translate((float(travel), 0, 0))
        assert (body ^ moved).volume() < 1e-6, travel
        assert (clip ^ moved).volume() < 1e-6, travel
        assert (stop ^ moved).volume() < 1e-6, travel
        assert (rail() ^ moved).volume() < 1e-6, travel
        # Tunnel walls, roof and floor prevent transverse escape at every position.
        for offset in [(0, -0.7, 0), (0, 0.7, 0), (0, 0, -0.7), (0, 0, 0.7)]:
            assert (clip ^ moved.translate(offset)).volume() > 0, (travel, offset)
    # Even at the extremes of available lateral/vertical play, the solid
    # tail cannot bypass the external screw during withdrawal.
    for dy in (-0.4, 0, 0.4):
        for dz in (-0.5, 0, 0.5):
            assert (stop ^ rigid.translate((6.7, dy, dz))).volume() > 0
    # Locking tooth really blocks uncommanded release; captive stop blocks overtravel.
    assert (body ^ slider.translate((0.6, 0, 0))).volume() > 0
    assert (clip ^ slider.translate((0.6, 0, 0))).volume() > 0
    assert (stop ^ rigid.translate((6.7, 0, 0))).volume() > 0
    assert (clip ^ rigid.translate((-0.5, 0, 0))).volume() > 0

    assembly = mesh(body + slider)
    assert len(assembly.split()) == 2
    print_extents = {}
    for name, part in parts:
        printable = mesh(part)
        printable.apply_translation(-printable.bounds[0])
        path = OUT / f'jetson-orin-din-{name}.stl'
        printable.export(path)
        saved = trimesh.load_mesh(path)
        assert saved.is_volume and len(saved.split()) == 1, name
        print_extents[name] = saved.extents.round(3).tolist()

    # Colour-separated geometry-only viewer: PCB envelope is a visual reference.
    scene = trimesh.Scene()
    for name, part, colour in [
        ('bracket', body, [60, 100, 140, 255]),
        ('separate-slider', slider, [240, 150, 35, 255]),
        ('M3-stop-screw-shaft-reference', short_edge(cyl(STOP_D / 2, 8, (STOP_X, STOP_Y, 3))),
         [190, 190, 195, 255]),
        ('DIN-rail-reference', rail(), [155, 160, 165, 255]),
        ('PCB-envelope-reference', box((79, 100, 1.57), (0, 0, PCB_Z)),
         [45, 130, 70, 130]),
    ]:
        tri = mesh(part)
        tri.visual.vertex_colors = colour
        tri.apply_scale(0.001)
        scene.add_geometry(tri, node_name=name)
    scene.export(OUT / 'assembly.glb')
    report = {
        'status': 'prototype; no physical print, fit, thermal or load testing',
        'units': 'mm (STL); metres (GLB)',
        'mount_edge': 'bottom 79 mm short edge, opposite GPIO',
        'slider_release_direction': '+X (right from fan side)',
        'assembly_extents_mm': assembly.extents.round(3).tolist(),
        'print_extents_mm': print_extents,
        'plastic_volume_cm3': round(assembly.volume / 1000, 2),
        'watertight_shells': 2,
        'shells': 'separately printed chassis and slide-in latch',
        'clip_test': 'complete standalone tunnel clip; uses the same slider and M3 screw',
        'clip_test_geometry_checks_passed': True,
        'stop_hardware': 'one M3 x 8 mm screw behind solid slider tail; 2.5 mm roof pilot',
        'retention': 'four-sided tunnel; external screw blocks withdrawal; rear shoulder stops insertion',
        'transverse_retention_checked': True,
        'withdrawal_stop_checked_at_clearance_extremes': True,
        'maximum_withdrawal_mm': 6.5,
        'vertical_slider_clearance_mm': GAP,
        'rigid_insertion_path_checked_mm': [0, 38],
        'insertion_samples': 381,
        'rigid_slider_travel_checked_mm': [0, 6.5],
        'travel_samples': 66,
        'nominal_rail_interference_mm3': 0,
        'positive_lock_and_end_stops_checked': True,
        'flexure_deformation_simulated': False,
        'pcb_holes_xy_mm': HOLES,
        'pilot_diameter_mm': PILOT_D,
        'pcb_underside_z_mm': PCB_Z,
    }
    (OUT / 'validation.json').write_text(json.dumps(report, indent=2) + '\n')
    print(json.dumps(report, indent=2))


if __name__ == '__main__':
    main()
