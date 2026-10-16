"""Generate millimetre-scale, closed STL meshes using only Python's standard library.

Run: python3 mechanical/printed-drums/generate.py
Profiles are (radius, height), revolved about Z. Mounting holes are intentionally
drilled after printing: the Nexus 18014 bolt pattern remains unverified.
"""
from pathlib import Path
import math
import struct

OUT = Path(__file__).parent
SEGMENTS = 256


def revolve(name, profile):
    vertices = [(r * math.cos(2 * math.pi * j / SEGMENTS),
                 r * math.sin(2 * math.pi * j / SEGMENTS), z)
                for r, z in profile for j in range(SEGMENTS)]
    triangles = []
    for i in range(len(profile)):
        for j in range(SEGMENTS):
            a = i * SEGMENTS + j
            b = i * SEGMENTS + (j + 1) % SEGMENTS
            c = ((i + 1) % len(profile)) * SEGMENTS + (j + 1) % SEGMENTS
            d = ((i + 1) % len(profile)) * SEGMENTS + j
            triangles.extend(((a, b, c), (a, c, d)))
    # Each edge must belong to exactly two faces; signed volume must be positive.
    edges = {}
    volume = 0.0
    records = []
    for face in triangles:
        a, b, c = [vertices[k] for k in face]
        u, v = [b[k] - a[k] for k in range(3)], [c[k] - a[k] for k in range(3)]
        n = (u[1]*v[2]-u[2]*v[1], u[2]*v[0]-u[0]*v[2], u[0]*v[1]-u[1]*v[0])
        length = math.sqrt(sum(x*x for x in n))
        assert length > 0
        volume += (a[0]*(b[1]*c[2]-b[2]*c[1]) +
                   a[1]*(b[2]*c[0]-b[0]*c[2]) +
                   a[2]*(b[0]*c[1]-b[1]*c[0])) / 6
        records.append(struct.pack('<12fH', *(x/length for x in n), *a, *b, *c, 0))
        for p, q in zip(face, face[1:] + face[:1]):
            edge = tuple(sorted((p, q)))
            edges[edge] = edges.get(edge, 0) + 1
    assert all(count == 2 for count in edges.values())
    assert volume > 0, 'Profile winding must face outward'
    with (OUT / name).open('wb') as f:
        f.write(b'DepthPilot prototype; units mm'.ljust(80, b'\0'))
        f.write(struct.pack('<I', len(triangles)))
        f.writelines(records)
    print(f'{name}: {len(triangles)} triangles; closed mesh; {volume/1000:.1f} cm3')


# Y: 92 mm core, 200 mm flanges, 60 mm assembled length.
# Flange cavity locates nominal 56.8 mm metal hub with 0.6 mm diametral clearance.
revolve('y-body.stl', [
    (7.5, 0), (99, 0), (100, 1), (100, 5), (99, 6),
    (48, 6), (46, 8), (46, 54), (36, 54), (36, 14),
    (28.7, 14), (28.7, 8), (7.5, 8),
])
# Insert spigot into body; translate this part +50 mm Z when assembling.
revolve('y-removable-flange.stl', [
    (30, 0), (35.8, 0), (35.8, 4), (99, 4), (100, 5),
    (100, 9), (99, 10), (30, 10),
])

# X: 88 mm rigid barrel, 108 mm flanges, 42 mm assembled length.
revolve('x-body.stl', [
    (7.5, 0), (53, 0), (54, 1), (54, 5), (53, 6),
    (46, 6), (44, 8), (44, 36), (34, 36), (34, 14),
    (28.7, 14), (28.7, 8), (7.5, 8),
])
# Assemble at +32 mm Z.
revolve('x-removable-flange.stl', [
    (30, 0), (33.8, 0), (33.8, 4), (53, 4), (54, 5),
    (54, 9), (53, 10), (30, 10),
])
# TPU trial sleeve: 0.4 mm diametral interference, fitted at +8 mm Z.
revolve('x-tpu-sleeve.stl', [
    (43.8, 0), (45.5, 0), (46, 0.5), (46, 27.5),
    (45.5, 28), (43.8, 28),
])


def template(axis, radius, diameter):
    """Full-scale drilling guide for six flange tie bolts, NOT the hub bolts."""
    size = diameter + 10
    centre = size / 2
    lines = [f'<svg xmlns="http://www.w3.org/2000/svg" width="{size}mm" height="{size}mm" viewBox="0 0 {size} {size}">',
             '<g fill="none" stroke="black" stroke-width="0.2">',
             f'<circle cx="{centre}" cy="{centre}" r="{diameter/2}"/>',
             f'<circle cx="{centre}" cy="{centre}" r="7.5"/>']
    for j in range(6):
        x = centre + radius * math.cos(j * math.pi / 3)
        y = centre + radius * math.sin(j * math.pi / 3)
        lines.append(f'<circle cx="{x}" cy="{y}" r="2.25"/>')
    lines.extend(['</g>', f'<text x="10" y="15" font-size="4">{axis}: six 4.5 mm holes, PCD {radius*2} mm</text>',
                  f'<path d="M 10 {size-10} h 50" stroke="black"/>',
                  f'<text x="10" y="{size-13}" font-size="4">50 mm scale check</text>', '</svg>'])
    (OUT / f'{axis.lower()}-flange-drill-template.svg').write_text('\n'.join(lines))


template('Y', 41, 200)
template('X', 39, 108)
