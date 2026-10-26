"""Generate identical, unloaded MG23-G20-D8 demonstration drums (millimetres).

Install requirements.txt and run this file. STL files are in print orientation.
"""
from pathlib import Path
import math

import manifold3d as m
import numpy as np
import trimesh

OUT = Path(__file__).parent
SEGMENTS = 192
BORE = 14.3  # Tune after printing shaft-fit-test.stl; nominal shaft is 14 mm.
CLAMP_HEIGHTS = (13, 24)


def cylinder(radius, height, z=0, segments=SEGMENTS):
    return m.Manifold.cylinder(height, radius, radius, segments).translate((0, 0, z))


def box(size, origin):
    return m.Manifold.cube(size).translate(origin)


def revolve(profile):
    return m.CrossSection([profile]).revolve(SEGMENTS)


def clamp(base):
    # Two lugs straddle a radial split; bolts squeeze the bore onto a bare shaft.
    solid = base + cylinder(17, 24, 6)
    solid += box((15, 24, 24), (8, -12, 6))
    solid -= cylinder(BORE / 2, 34, -1)
    # Lead-in at the gearbox-facing end.
    solid -= m.Manifold.cylinder(0.8, BORE / 2 + 0.6, BORE / 2, SEGMENTS)
    solid -= box((24, 1.6, 32), (0, -0.8, -1))
    for z in CLAMP_HEIGHTS:
        hole = cylinder(2.25, 40).rotate((90, 0, 0)).translate((18, 20, z))
        solid -= hole
        # M4 hex nut pocket: 7.3 mm across flats, open at +Y lug face.
        pocket = cylinder(7.3 / math.sqrt(3), 3.6, segments=6)
        solid -= pocket.rotate((90, 0, 0)).translate((18, 12.1, z))
    return solid


def save(name, solid):
    assert solid.status() == m.Error.NoError, (name, solid.status())
    mesh = solid.to_mesh()
    tri = trimesh.Trimesh(np.asarray(mesh.vert_properties)[:, :3],
                          np.asarray(mesh.tri_verts), process=True)
    assert tri.is_watertight and tri.is_winding_consistent and tri.volume > 0, name
    assert len(solid.decompose()) == 1, f"Disconnected geometry: {name}"
    tri.export(OUT / name)
    # Validate the exported STL, including merged coincident vertices.
    check = trimesh.load_mesh(OUT / name)
    assert check.is_watertight and check.is_volume, name
    print(f"{name}: {len(tri.faces)} triangles; {tri.volume / 1000:.1f} cm³; "
          f"bounds {tri.extents.round(2).tolist()} mm")


def main():
    # Lower flange and hollow barrel: 80 OD / 64 ID, 40 clear winding width.
    # 0.8 mm flange edge chamfers; a 1 mm internal rope-root chamfer.
    body = revolve([(0, 0), (54.2, 0), (55, 0.8), (55, 5.2),
                    (54.2, 6), (41, 6), (40, 7), (40, 46),
                    (32, 46), (32, 6), (0, 6)])
    body = clamp(body)
    for z in CLAMP_HEIGHTS:
        # Straight driver access from -Y, coaxial with the clamp bolt heads.
        # Stop in the cavity, short of the lugs; 6 mm throat for a 3 mm hex key.
        port = cylinder(3, 40).rotate((90, 0, 0)).translate((18, -20, z))
        # Flared entry intersects the curved barrel, softening the outer lip.
        entry = m.Manifold.cylinder(11, 3, 6.7, SEGMENTS)
        port += entry.rotate((90, 0, 0)).translate((18, -31, z))
        body -= port
    # Print outer flange flat, spigot pointing up. Flip for assembly.
    lid = revolve([(0, 0), (54.2, 0), (55, 0.8), (55, 3.2),
                   (54.2, 4), (31.7, 4), (31.7, 6.5),
                   (31.2, 7), (0, 7)])
    for angle in (45, 135, 225, 315):
        x, y = (35 * math.cos(math.radians(angle)),
                35 * math.sin(math.radians(angle)))
        body -= cylinder(2.25, 48, -1).translate((x, y, 0))
        body -= cylinder(7.3 / math.sqrt(3), 4.1, -0.1, 6).translate((x, y, 0))
        lid -= cylinder(2.25, 9, -1).translate((x, y, 0))
    for z in CLAMP_HEIGHTS:
        # A 4 mm diameter tool envelope clears the wall up to the screw head
        # (M4 socket head + 1 mm washer puts its outer face at Y=-17).
        tool = cylinder(2, 43).rotate((90, 0, 0)).translate((18, -17, z))
        assert (body ^ tool).volume() < 0.001, 'Clamp driver access obstructed'
    save('drum-body.stl', body)
    save('outer-flange.stl', lid)
    save('shaft-fit-test.stl', clamp(cylinder(26, 6)))
    # Geometric interference check in assembled coordinates (50 mm overall).
    assembled_lid = lid.rotate((180, 0, 0)).translate((0, 0, 50))
    assert (body ^ assembled_lid).volume() < 0.001, 'Flange interferes with body'


if __name__ == '__main__':
    main()
