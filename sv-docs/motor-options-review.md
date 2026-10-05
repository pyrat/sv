# Winch motor options review

Reviewed: 5 October 2026. This is a procurement shortlist, not a validated replacement specification. Prices are observations from public pages, not quotations.

## Recommendation

**Updated requirement:** the camera weighs **15 kg in air**, but the winch only needs to reposition it underwater; lifting it out of the water is not required. The user wants some operating headroom. Submerged weight, tether drag and full-drum radius remain unknown, so dry weight alone cannot establish required torque. The original 16 Nm rating is a comparison point, not a minimum.

**Updated speed target:** approximately **5 seconds per metre (0.20 m/s)**. Assuming direct drive of the documented 92 mm barrel, this requires approximately **42 rpm** at the first layer. Seek **10–12 Nm continuous at roughly 42–45 rpm**, with controllable speed. This remains a provisional torque target, not validated sizing: submerged line tension and maximum winding radius are still needed. Retain an 8 Nm option for comparison once actual tension is known.

The first enquiry remains **Twirl**, but ask them to select the motor family rather than presuming PG52ZY52 is sufficient. At 42 rpm, 10–12 Nm requires approximately **44–53 W at the gearbox output**. With illustrative 70–80% gearbox efficiency, that requires approximately **55–75 W mechanical output from the bare motor**, before additional margin. A roughly **80–100 W bare-motor mechanical-output class** is a useful enquiry starting point; the 30–60 W PG52 family may be marginal at the upper target. Do not confuse mechanical output with electrical input power.

At 46 mm winding radius, 10–12 Nm corresponds to an ideal 217–261 N line pull. At 90 mm radius it provides only 111–133 N, while approximately 21 rpm is sufficient for 0.20 m/s. These conversions do not establish an air-lifting rating or guaranteed operating margin. At constant rpm, line speed increases as rope builds up; approximately constant linear speed requires adjusting rpm for winding radius. If five seconds includes acceleration and stopping over a one-metre move, the cruising speed must exceed 0.20 m/s.

For a more fully specified industrial option, consider **Parvalux BRx70-40 / GB12, 24 V, 60:1**, part **781093-735902**, with brake and encoder: 7.5 Nm continuous at 50 rpm. Its public price is high, so lower motor power does not necessarily mean a cheaper assembly.

The larger Dunkermotoren/Parvalux options remain below as full-performance comparisons. The lower-cost STEPPERONLINE options require additional feedback/brake integration. All suggestions assume the motor stays above water or inside a suitable dry enclosure; a submerged motor needs a separately specified underwater assembly.

## Smaller-motor shortlist for the camera

### A. Twirl PG52ZY52 with brake and encoder — first supplier enquiry

- A different named manufacturer: Ningbo Twirl, rather than Smart Automation.
- 52 mm brushed-DC planetary gearmotor family, 12/24 V; manufacturer range page lists 30–60 W.
- Dedicated assembly page documents an encoder and a power-off electromagnetic brake, with a stated brake rating of 2 Nm and 9.7 W coil power. Confirm which shaft the brake torque rating refers to.
- Family headline advertises 50–300 kg·cm (about 4.9–29.4 Nm). This is a range across configurations, not a claim that every ratio can supply that torque at every speed.
- Under the latest camera requirement, ask for **10–12 Nm continuous at roughly 42–45 rpm**, and an 8 Nm comparison option. The vendor must confirm a feasible motor/gearbox combination and whether a larger family than PG52ZY52 is needed. At 42 rpm, 12 Nm is about 53 W at the output, before allowing for gearbox losses in the motor selection.
- The page contradicts itself on encoder resolution: headline 13 PPR, table 3 PPR. Obtain the selected encoder datasheet, channel arrangement and output voltage.
- Price/MOQ require quotation. Page advertises 1–2 weeks for samples and 3–4 weeks for production; confirm current timing.
- Contact: **sales@nbtwirl.com**.

Sources: [brake/encoder assembly](https://www.twirlmotor.com/52mm-dc-gear-motor-with-brake-and-encoder/), [motor family range](https://www.twirlmotor.com/planet-gear-dc-motor/).

### B. Parvalux BRx70-40 / GB12, 24 V, 60:1 — documented industrial option

- Base part **781093-735902**; **7.5 Nm continuous at 50 rpm**, 12.6 Nm intermittent (S2, 15 minutes).
- 79 W brushed motor, **4.7 A continuous**; approximately 39 W at the gearbox output operating point, with 50% listed gearbox efficiency.
- Right-angle layout, published base envelope **250 × 81 × 138 mm**; requires different mounting from the reference inline assembly.
- Brake and incremental encoder available through modular configuration; get the complete assembly code and power-off brake specification.
- Public price observed: **£925.31 each for 1–4**, **£693.98 each for 5–10**, **On Request**, before accessory additions and any shipping/taxes.
- The page's descriptive text says composite, whereas its title and specification table say bronze. Confirm the material in the quote.

Source: [product, configuration and enquiry](https://www.parvalux.com/us/product/brx70-40-24v-3000rpm-gb12-601-bronze/).

### C. STEPPERONLINE F5B60-24GN-30S / 5GN50K — low-cost prototype lead

The indexed vendor listing specifies **24 V, 60 W, 6.3 Nm at 60 rpm**, with a 50:1 spur gearbox and 90 × 90 mm flange. The search listing showed **US$41.44** and stock, but the full page could not be independently retrieved during this follow-up, so confirm price, included components and availability directly. It is BLDC and therefore needs a different drive from the archived brushed H-bridges. A holding brake and compatible quadrature encoder have not been established for this SKU.

Source: [product listing](https://www.omc-stepperonline.com/24v-60w-60rpm-geared-brushless-dc-motor-6-30nm-892-16oz-in-50-1-spur-gearbox-f5b60-24gn-30s-5gn50k).

### Camera-load sizing examples

At a 50 mm effective winding radius, 3 / 5 / 8 Nm correspond to ideal line pulls of 60 / 100 / 160 N, respectively (about 6.1 / 10.2 / 16.3 kg-force). At 100 mm radius, those forces halve. At 50 mm radius, 30–75 rpm gives approximately 0.16–0.39 m/s line speed.

These are mechanical conversions before allowance for losses and transients, not rated payloads. For the camera, establish submerged weight, suspended cable weight, tether tension and current-induced drag. If the system also retrieves the camera out of the water, include that air-weight load case. A near-neutrally buoyant camera can still produce appreciable tether force in a current.

## Archive coverage

The original checkout contains 179 tracked files. Review included the complete file inventory, a repository-wide text scan for motor sizing information, the motor/drum/main-board/driver PDFs, all three main PNG drawings, operating notes, the Draw.io labels, and targeted motor/control source inspection. The two ZIP bundles were inventoried and searched for additional motor specifications. This was a motor-selection review, not a line-by-line firmware audit or executable reverse-engineering exercise.

| Material | Relevance to motor selection |
| --- | --- |
| `dawinchi/56jxe300k.63zyt03b-24v.brake+hs30b_-50p-8a-v_外形图&nbsp_180202a-1_3959_4626.pdf` | Primary motor assembly specification; drawing 180202A-1, revision A, 29 January 2023, status Development. |
| `dawinchi/drum-115669.pdf` | Hoisting pulley geometry, 12 mm bore, 92 mm winding barrel and 200 mm flanges. |
| `dawinchi/8610-0533_rev_1.0.0_shts_1_-_5_of_5.pdf` | 24 V supply, two brushed motor drivers, separate brake switching, encoder counters and regenerative-energy shunt. |
| `ropey/motor_driver_4020593.pdf` | DFRobot DFR0601 dual brushed-DC driver; a different driver from the daWinchi main board. |
| `dawinchi/CCP-wire-diag-1.png`, `CCP-wire-diag-2.png` | Power switching, emergency-stop power removal, and supply current sensing. |
| `dawinchi/winch-move-algo.txt`, `NTNU-site-test.txt` | Installation-specific positions and current observations; no defensible payload/duty specification. |
| Archived Go/TypeScript firmware and tests | Separate brake control, encoder-based moves, ramps and calibrated two-axis relationships. |
| `dawinchi/wiggle/`, `ropey/mt/` | Repeated-motion client and encoder experiment; neither establishes motor load ratings. |
| READMEs, estimate documents, Draw.io diagram, button notes, panel script and OLED artwork | Software/project context; no additional motor sizing specification identified. |
| Adafruit sources/ZIPs, INA226 library, build files, dependency locks, generated API documents and bundled binaries | Supporting software rather than motor selection data. Binaries were identified, not run. |

Paths in this table are relative to `sv-docs/`.

## Reference assembly

Manufacturer named on the drawing: **Smart Automation**.

Assembly: **56JXE300K.63ZYT03B-24V.brake+HS30B-50P-8A-V**.

The visible specification table gives:

| Parameter | Reference value |
| --- | --- |
| Supply | 24 V DC |
| Motor technology | Brushed DC, supported by motor family and two-wire H-bridge drive arrangement |
| Reduction | 55:1 |
| Gearmotor rated output torque | 16 Nm |
| Gearmotor rated output speed | 72.7 rpm |
| Gearmotor rated mechanical output | 122.3 W |
| Gearmotor no-load speed/current | 84.6 rpm / 1 A |
| Rated current | 9.6 A |
| Bare motor rated torque/speed/power | 0.4 Nm / 4,000 rpm / 167.6 W |
| Brake | WR220, 24 V, 1.0 Nm, mounted on motor side |
| Encoder | HS30B-50P; 50 pulses/revolution is indicated by the designation and requires confirmation |
| Enclosure/insulation | IP50; class F, 155°C insulation class |
| Gearbox/motor diameter | 56 mm / 63 mm |
| Output shaft | Stainless, 12 mm diameter, 28 mm projection, 4 mm keyway, end M4 tapped 10 mm deep |
| Mounting | Four M5 holes on 45 mm pitch circle; 32 mm pilot, 2 mm projection |
| Approximate overall axial length | 259 mm including output shaft, summing the drawing's dimension chain |

The PDF text extraction also exposes conflicting specification fragments (including 3,000 rpm and 7.5 A, and other current limits) which do not appear in the rendered final table. The visible table is the baseline used here; the drawing's Development status and conflicting embedded text mean the as-built specification should be confirmed rather than mixing those values.

The 1 Nm brake rating is **not** the geared output torque rating, and multiplying it by 55 does not establish the assembly's permissible holding load. Gearbox static limits, reverse transmission characteristics and brake duty must be checked separately.

### What the drum implies

For a directly driven drum, with line tension F, winding radius r and speed n:

```
Required drum torque = F × r
Line speed = 2 × pi × r × n / 60
```

At a nominal 46 mm barrel radius, 16 Nm corresponds to approximately **348 N** of ideal line tension (35.5 kg-force), and 72.7 rpm corresponds to **0.35 m/s**. At an illustrative 90 mm rope-layer radius, those values become **178 N** and **0.69 m/s**. Neither figure is a safe working load: rope diameter, actual full-drum radius, reeving, friction, acceleration and service factors remain unspecified. The 200 mm flange diameter is not the empty-drum winding diameter.

## Full-reference-performance comparisons

### 1. Dunkermotoren GR 80x80, 24 V, with PLG gearbox, brake and encoder

**Best first enquiry for a different manufacturer's inline assembly.**

The published 24 V bare-motor data are 195 W continuous, 0.59 Nm at 3,150 rpm, 10 A nominal current and 4 kg motor weight. Its modular options include PLG 63 EP/HT and PLG 75 EP/HT planetary gearboxes, E 90 R / E 100 R/A brakes, and RE-series encoders. The supplier must confirm the simultaneous brake/encoder build and select the power-off brake variant.

Illustrative sizing, not a catalogued assembly rating: 3,150 / 73 suggests a ratio near 43:1. At an assumed 80% gearbox efficiency, 0.59 × 43 × 0.8 gives about 20.3 Nm. The actual available ratio, continuous gearbox rating and torque-speed curve must be confirmed.

Important electrical difference: the datasheet gives **112 A starting current** for the 24 V motor. A 10 A nominal current does not make it a plug-in replacement for either archived controller. Specify a suitably current-limited drive and verify starting torque and supply capacity.

The smaller GR 63x55 24 V model is listed at 0.27 Nm at 3,350 rpm, approximately 95 W before gearing; GR 63Sx55 24 V is about 117 W before gearing. Neither meets the reference's 122 W geared mechanical output continuously merely by changing the reduction ratio, but both are legitimate lower-power candidates under the updated camera requirement.

Ordering: manufacturer/distributor quotation; no verified complete-assembly price or lead time.

Sources: [GR80 product and enquiry page](https://www.powerbox.com.au/products/dunkermotoren-gr80-series), [manufacturer datasheet hosted by distributor](https://www.powerbox.com.au/cdn/shop/files/GR_80_x40_x80_Brushed_DC_Motor_Datasheet.pdf), [GR63 comparison](https://www.powerbox.com.au/products/dunkermotoren-gr63-series).

### 2. Parvalux BRx90-75 / GB9, 24 V, 30:1

**Concrete industrial combination with documented brake/encoder accessory route.**

- Base part: **781107-735895**.
- Published output: **20.3 Nm continuous at 100 rpm**, 33.8 Nm intermittent (S2, 15 minutes).
- Motor: 283 W, 24 V, 14.5 A continuous; 283 W is motor power, not geared-shaft power. The published output operating point represents about 213 W.
- Right-angle worm-wheel gearbox, 75% listed efficiency, composite gear version.
- Product page dimensions: **333 × 90 × 180 mm**; listed combination weight 6.4 kg. Confirm final dimensions including accessories; older range-guide dimensions differ.
- Modular guide lists a **1.5 Nm motor-side brake** and **48–2,048 PPR incremental encoder** options for BRx90. Confirm power-off operation, brake voltage, electrical interface and the complete ordering code.
- Public page showed **£891.15 each for 1–4**, **£668.36 each for 5–10**, availability **On Request**. These are base-combination prices; brake, encoder, shaft kit, shipping and applicable taxes need quotation.

It could be controlled down toward 73 rpm, subject to supplier confirmation of thermal/torque capability there. Its 14.5 A nominal rating exceeds the Ropey DFR0601's 12 A continuous channel rating. Mechanical layout, controller and power budget need rechecking.

The product page has a stray “Bronze” description although its title and specification table say “Composite”; identify the gear material explicitly in the quotation.

Sources: [product/configuration/enquiry page](https://www.parvalux.com/us/product/brx90-75-24v-3000rpm-gb9-301-composite/), [combination datasheet](https://www.parvalux.com/product/brx90-75-24v-3000rpm-gb9-301-composite/datasheet.pdf), [modular guide, especially pages 19–21](https://www.parvalux.com/app/uploads/2024/02/Parvalux-Modular-Range-Guide-2024.pdf).

### 3. STEPPERONLINE M82B200 planetary BLDC family

**Lower-cost development option, requiring substantial integration.**

The directly listed **M82B200-24P-30S/G82-30S1** is 24 V, 10.5 A rated, **15 Nm at 100 rpm**, with 82 mm frame and 19 mm keyed output shaft. The listing showed **US$58.48**, with stock indicated, before shipping/taxes. This 30:1 model falls short of the 16 Nm baseline.

The same product page's ratio table lists **35.1:1 = 18 Nm at 85 rpm** and **46.2:1 = 24 Nm at 65 rpm**. Ask the vendor for the exact ordering code and availability of those variants; the observed price and stock belong to the 30:1 SKU only.

The documented feedback is Hall commutation sensing, not the existing A/B/index quadrature encoder. No holding brake is specified. It needs a BLDC controller, a supported brake solution and suitable position feedback. The listed 120 N radial-load limit on the 30:1 assembly is also a reason to consider an independently supported drum rather than putting rope load directly on the output shaft; confirm limits for the selected variant.

Source: [product page and ratio table](https://www.omc-stepperonline.com/24v-200w-100rpm-geared-brushless-dc-motor-15-00nm-2124-17oz-in-30-1-planetary-gearbox-m82b200-24p-30s-g82-30s1).

### 4. Brushless.com BLDC-GEAR-P200, 24 V, 40:1, brake option

**Additional online configurable BLDC lead, less completely specified than the industrial assemblies.**

The listing offers motor model **D5BLD200-24A-30S**, 200 W, 0.64 Nm at 3,000 rpm, 10.42 A, a 40:1 planetary gearbox and an optional 24 V brake. That ratio gives a calculated nominal 75 rpm. At an assumed 80% efficiency the calculated output is about 20.5 Nm, but the listing does not establish the assembled gearbox's continuous torque limit, efficiency or holding rating.

Observed displayed component prices total **US$1,018.11**: base $530.62 + 40:1 option $47.53 + brake $206.44 + 24 V controller $233.52. This is arithmetic from advertised options, not a verified checkout quote. The page advertises 6–25 day delivery; destination-specific timing remains unverified.

Encoder provision, brake holding torque and power-off behavior, complete assembly drawing, gearbox load limits, and actual OEM identity need confirmation. It is a different product listing, but independent manufacture from the reference supplier has not been established.

Source: [configurable listing](https://www.brushless.com/200w-bldc-planetary-gear-motor).

## Integration findings

- **Motor drive:** daWinchi's schematic uses two Pololu G2 18v17 H-bridge modules. Ropey's separate PDF describes DFR0601, 6.5–37 V, 12 A continuous per channel and 70 A for 100 ms. Both are brushed-motor drives; neither directly commutates a three-phase BLDC motor.
- **Holding brake:** `dawinchi/old-repo/dawinchi/dawinchi-hw/internal/motor/break.go` drives the brake pin high to release and low to engage. The schematic switches a separate brake coil. Specify spring-applied, electrically released behavior for a comparable holding function, and check coil current and timing.
- **Encoder:** `internal/ls7366r/ls7366r.go` configures x4 quadrature counting and describes HS30B A/B TTL outputs with index. The board supplies encoder power at 5 V but uses 3.3 V counter circuitry and pull-ups. Confirm output topology and voltage compatibility rather than treating any 5 V encoder as interchangeable.
- **Calibration:** if the reference encoder really is 50 PPR at the motor shaft, 50 × 4 × 55 = **11,000 counts/output revolution**. A new gear ratio/PPR changes this value. Stored positions, travel limits, ramps and two-axis coordination require recalibration; electrical pulse resolution is not mechanical positioning accuracy.
- **Regeneration:** the daWinchi power sheet includes a shunt for energy fed back by motors. A replacement drive must handle lowering/deceleration energy as well as motoring current.
- **Environment:** IP50 does not provide water protection. Confirm whether the motor is inside a dry sealed enclosure or needs its own wet/salt-exposure protection. A stainless shaft alone does not establish marine suitability.
- **Field currents:** `winch-move-algo.txt` records differing meter/log current values. These are not a substitute for measured worst-case motor current and duty cycle.

## Draft supplier enquiry

> Please quote a complete geared motor assembly for a reversible 24 V DC winch repositioning an underwater camera weighing 15 kg in air. Air retrieval is not required, but we would like operating headroom. Desired line speed is approximately 0.20 m/s (5 seconds per metre). Assuming a directly driven 92 mm winding barrel, our provisional output target is 10–12 Nm continuous at roughly 42–45 rpm; please also quote an 8 Nm option and state the available torque-speed combinations. Please select an adequately rated motor family, allowing for gearbox losses. We would like a spring-applied/electrically released 24 V holding brake and incremental quadrature position feedback. Please state the continuous and intermittent torque-speed ratings, duty assumptions, gearbox efficiency and permitted radial/axial loads; brake holding torque and switching times; encoder PPR, supply, output topology and pinout; motor nominal/start/stall current and recommended drive; and the assembly dimensions, mass and environmental rating. Please include the complete ordering code, sample pricing, quantity pricing and lead time. Final sizing will be confirmed against our submerged camera/tether line pull, full-drum radius, line speed and operating cycle.

Add the actual quantities, delivery country, budget, operating environment and mechanical constraints before sending. For a new design, let the supplier propose its standard mounting/shaft rather than prescribing every reference-drawing dimension.

## Information needed to choose the order

1. Maximum line pull or payload, reeving arrangement, drum winding radii and required line speed.
2. Longest run, cycles per hour, holding duration and ambient temperature.
3. Whether the motor must fit an existing enclosure/drum, or mounts and coupling can change.
4. Whether the existing controller must be reused or a new drive is acceptable.
5. Quantity, budget per complete assembly, delivery country and required date.

## Distinct motor selection and provenance

A different manufacturer's standard motor assembly gives a genuinely different component selection. Keep the selection rationale grounded in the new winch's independently established performance requirements and the supplier's public specifications.

Changing the motor does not establish freedom to reuse the surrounding design. This archive's Imenco drum and electrical drawings carry explicit copying/use restrictions, and the electrical drawings are marked commercial in confidence. Do not treat a motor substitution as resolving those separate rights. A supplier enquiry can describe your own performance requirements without forwarding those drawings.
