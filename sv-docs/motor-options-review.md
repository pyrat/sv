# Winch motor options review

Reviewed: 5 October 2026. This is a procurement shortlist, not a validated replacement specification. Prices are observations from public pages, not quotations.

## Recommendation

**Updated requirement:** the camera weighs **15 kg in air**, but the winch only needs to reposition it underwater; lifting it out of the water is not required. The user wants some operating headroom. Submerged weight, tether drag and full-drum radius remain unknown, so dry weight alone cannot establish required torque. The original 16 Nm rating is a comparison point, not a minimum.

**Updated speed target:** approximately **5 seconds per metre (0.20 m/s)**. Assuming direct drive of the documented 92 mm barrel, this requires approximately **42 rpm** at the first layer. Seek **10–12 Nm continuous at roughly 42–45 rpm**, with controllable speed. This remains a provisional torque target, not validated sizing: submerged line tension and maximum winding radius are still needed. Retain an 8 Nm option for comparison once actual tension is known.

The first enquiry remains **Twirl**, but ask them to select the motor family rather than presuming PG52ZY52 is sufficient. At 42 rpm, 10–12 Nm requires approximately **44–53 W at the gearbox output**. With illustrative 70–80% gearbox efficiency, that requires approximately **55–75 W mechanical output from the bare motor**, before additional margin. A roughly **80–100 W bare-motor mechanical-output class** is a useful enquiry starting point; the 30–60 W PG52 family may be marginal at the upper target. Do not confuse mechanical output with electrical input power.

At 46 mm winding radius, 10–12 Nm corresponds to an ideal 217–261 N line pull. At 90 mm radius it provides only 111–133 N, while approximately 21 rpm is sufficient for 0.20 m/s. These conversions do not establish an air-lifting rating or guaranteed operating margin. At constant rpm, line speed increases as rope builds up; approximately constant linear speed requires adjusting rpm for winding radius. If five seconds includes acceleration and stopping over a one-metre move, the cruising speed must exceed 0.20 m/s.

For a more fully specified industrial option, consider **Parvalux BRx70-40 / GB12, 24 V, 60:1**, part **781093-735902**, with brake and encoder: 7.5 Nm continuous at 50 rpm. Its public price is high, so lower motor power does not necessarily mean a cheaper assembly.

The larger Dunkermotoren/Parvalux options remain below as full-performance comparisons. The lower-cost STEPPERONLINE options require additional feedback/brake integration. All suggestions assume the motor stays above water or inside a suitable dry enclosure; a submerged motor needs a separately specified underwater assembly.

## Leadshine brake stepper + STEPPERONLINE gearbox (8 October 2026)

The user proposed the [DamenCNC Leadshine CS-M22323B](https://www.damencnc.com/en/closed-loop-stepper-motor-with-brake-nema23-2-3nm-cs-m22323b-2phase/a5340). Its drawing specifies an 8 mm D-shaft projecting 21 ±1 mm, a 38.1 mm locating pilot, 47.14 mm square mounting-hole spacing, and a 57.15 mm NEMA23 frame. It has 2.3 Nm holding torque, 5 A phase current, a 1000-line encoder, and a 2 Nm power-off brake released with 24 V (4.5 W). Listed motor price is €140 excluding VAT; the recommended CS-D507E driver is listed at €80 excluding VAT.

These separate gearboxes are plausible motor-side matches based on their published NEMA23 interface, 8 mm input and accepted shaft lengths. The user subsequently supplied the EG23-G20-D8 drawing A1783, revision 0, dated 3.8.2024. This confirms the principal motor-side dimensional match with CS-M22323B: Ø8 G6 input bore, accepted shaft projection 12–31 mm versus the motor's 21 ±1 mm, Ø38.1 H7 pilot recess versus the motor's Ø38.1 +0/−0.05 mm pilot, and 47.14 mm square mounting pattern. The gearbox has four M4 mounting threads, 10 mm deep; the motor has Ø5 mm through holes. The 4.5 mm pilot recess accommodates the motor's 1.6 mm pilot projection. No motor-side adapter or shaft shortening is indicated by these dimensions. Select mounting screw length to suit the motor flange and usable thread engagement without bottoming; confirm the input coupling's tightening instructions and suitability for the motor D-flat. The EG10 and MGL drawings have not been separately verified.

The supplied EG20 drawing also confirms 200 N maximum axial load, 400 N maximum radial load, 3,000 rpm rated input speed and 6,000 rpm maximum input speed. The required approximately 840 rpm input is within the gearbox speed rating, but motor running torque remains to be verified. The drawing does not specify the load application distance or combined-load conditions for those bearing limits, so they do not by themselves validate an overhung drum installation.

The user also supplied MGL23-G10-D8 drawing A4174, revision 0, dated 2.17.2025. Its principal motor interface matches: Ø8 +0.005/+0.015 mm input bore, 20–25 mm accepted shaft projection (motor 20–22 mm including tolerance), Ø38.1 +0.005/+0.03 mm locating recess, 47.14 mm square mounting pattern and four M4 threads 10 mm deep. Pilot recess depth is not explicitly dimensioned in this drawing; verify clearance for the motor's 1.6 mm pilot and follow the coupling instructions for the D-flat. No shaft shortening or adapter is indicated by the principal dimensions.

MGL10 is a credible lower-cost prototype option if the measured duty stays comfortably below its 10 Nm permissible torque limit. At approximately 420 motor rpm it gives the desired approximately 42 drum rpm; with 90% efficiency, 8 Nm output requires approximately 0.89 Nm motor torque and 10 Nm requires approximately 1.11 Nm, before acceleration reserve. Verify these points against the motor/driver torque-speed curve. Its 20 Nm momentary rating does not support sustained 12 Nm operation. The drawing confirms 1,000 rpm rated / 2,000 rpm maximum input speed, 100 N axial / 300 N radial load limits, and 66.3 mm body length (23.2 mm shorter than EG20). At 46 mm winding radius, 8–10 Nm gives approximately 174–217 N ideal line pull; at 90 mm, approximately 89–111 N. Actual headroom depends on submerged camera/tether tension. The motor can potentially overtorque this gearbox at low speed, so configured torque/current limits matter. The output remains 14 mm keyed and requires adaptation to the archived 12 mm drum bore.

| Gearbox | Ratio | Accepted input shaft length | Published permissible / momentary output torque | Efficiency | Observed USD price |
|---|---:|---:|---:|---:|---:|
| [EG23-G10-D8](https://www.omc-stepperonline.com/eg-series-planetary-gearbox-gear-ratio-10-1-backlash-15-arc-min-for-8mm-shaft-nema-23-stepper-motor-eg23-g10-d8) | 10:1 | 12–31 mm | 10 / 20 Nm | 94% | $34.09 |
| [MGL23-G10-D8](https://www.omc-stepperonline.com/mgl-series-planetary-gearbox-gear-ratio-10-1-backlash-30arc-min-for-8mm-shaft-standard-nema-23-stepper-motor-mgl23-g10-d8) | 10:1 | 20–25 mm | 10 / 20 Nm | 90% | $19.65 |
| [EG23-G20-D8](https://www.omc-stepperonline.com/eg-series-planetary-gearbox-gear-ratio-20-1-backlash-20-arc-min-for-8mm-shaft-nema-23-stepper-motor-eg23-g20-d8) | 20:1 | 12–31 mm | 30 / 60 Nm | 94% | $46.87 |

For the provisional 10–12 Nm target, **EG23-G20-D8 is the preferred candidate to validate**: the 10:1 products have a 10 Nm permissible output limit, so their 20 Nm momentary figure does not establish suitability for sustained 12 Nm operation. EG uses alloy-steel gears; MGL uses powder-metallurgy gears. Published no-load backlash is 15 arcmin for EG10, 20 for EG20 and 30 for MGL10.

At approximately 42 output rpm (0.20 m/s on the bare 92 mm barrel), 10:1 requires approximately 420 motor rpm and 20:1 approximately 840 rpm. Using the listed efficiencies, delivering 10 Nm through EG10 requires approximately 1.06 Nm running motor torque; delivering 12 Nm through EG20 requires approximately 0.64 Nm at 840 rpm. These exclude acceleration reserve. The 2.3 Nm holding rating does not prove either operating point. Obtain the CS-M22323B torque-speed curve for the selected driver, current and supply voltage; a 48 V driver supply is worth evaluating for high-speed torque, with separate 24 V brake release power. No suitable loaded torque-speed curve was verified in this review.

All three have 14 mm keyed output shafts (35 mm projection on EG, 34 mm on MGL), so they do not directly fit the archived 12 mm drum bore. A separately supported drum shaft and suitable coupling is a practical integration route. Check output bearing loads and overhang if supporting the drum from the gearbox. Gearbox IP54 is not an underwater rating.

The existing brushed H-bridges must be replaced by a compatible closed-loop stepper driver and step/direction control. Configure torque/current limits to respect the gearbox, and sequence brake release/engagement with motor holding torque. Do not treat the brake's torque multiplied by the reduction as an unrestricted output holding rating: gearbox static limits also apply. Motor encoder feedback does not measure gearbox backlash or changing drum radius.

## Leadshine PLF060-10-P2-BJ-8-38.1 follow-up

**User quotation:** €190 excluding VAT for the exact -8-38.1 variant. At this price, no verified application benefit justifies prioritising it over the STEPPERONLINE candidates. Deprioritise Leadshine PLF for this camera winch: favour EG23-G20-D8 for the provisional 10–12 Nm target, or MGL23-G10-D8 for a lower-cost prototype with measured duty comfortably below 10 Nm. Motor running torque at the selected speed still requires verification. STEPPERONLINE prices below are observed USD listings before shipping/taxes, not landed-cost quotations.

The exact model appears in [Ocean Controls' Leadshine model list](https://oceancontrols.com.au/images/content/MOG-001_Models.jpg), described as **10:1, for NEMA23 (57 mm) stepper motors with an 8 mm shaft**. [Leadshine's official PLF page](https://www.leadshine.com/products/other-product/gearboxes/plf-series.html) confirms straight-tooth planetary gearboxes for NEMA23/24/34 motors, but did not expose a detailed model datasheet. The exact -38.1 drawing, torque rating and current price were not verified.

Important variant distinction: [Ocean Controls MOG-001](https://oceancontrols.com.au/MOG-001.html) currently sells **PLF060-10-P2-BJ-8-36**, for NEMA24. Its [Leadshine-branded datasheet](https://oceancontrols.com.au/files/datasheet/lea/MOG-001.pdf) specifies a 36 mm pilot and a 70.7 mm mounting bolt-circle diameter (approximately 50 mm square), which do not match CS-M22323B's 38.1 mm pilot and 47.14 mm square pattern. Do not substitute this version for -38.1.

For reference only, the **-36 sibling's** sheet lists 96% full-load efficiency, ≤12 arcmin backlash, 3,000/6,000 rpm rated/maximum input speed, 485/380 N maximum radial/axial output load, IP65, 1.0 kg, 14 mm keyed output and 20,000 h average life. It prints **85 Nm nominal / 170 Nm emergency-stop torque**. These are not verified ratings for the requested -38.1. The nominal torque is unusually high compared with the other 60 mm units reviewed and needs explicit manufacturer confirmation; no corrected value is assumed. Other suppliers use PLF060 for different gearboxes, so their 16 Nm or 25 Nm specifications cannot establish this Leadshine model's rating either.

Comparison: MGL10 has 10 Nm permissible / 20 Nm momentary torque, 90% efficiency and ≤30 arcmin backlash; EG10 has 10/20 Nm, 94% and ≤15 arcmin; EG20 has 30/60 Nm, 94% and ≤20 arcmin. Leadshine's exact model offers the same 10:1 ratio as MGL10/EG10, requiring approximately 420 motor rpm for the target line speed. If the sibling's 96% efficiency also applies, 12 Nm output requires approximately 1.25 Nm motor torque at that speed, plus acceleration reserve. EG20 needs approximately 0.64 Nm at 840 rpm instead. Compare both points on the actual motor curve rather than assuming one ratio is inherently better.

Procurement conclusion: the Leadshine -38.1 is a credible NEMA23 candidate, but cannot yet be ranked above EG20 for torque capacity. Request its exact drawing (including accepted shaft length and pilot depth), continuous and transient torque ratings, confirmed CS-M22323B compatibility, price and lead time. MGL remains the lowest-cost candidate for a duty comfortably below 10 Nm; EG20 remains the better-documented candidate for the provisional 10–12 Nm target. IP65, if confirmed for the requested variant, is better dust/water-jet protection than IP54 but is not a submersion rating.

## STEPPERONLINE motor replacement with CS2RS-D507 Modbus drive

The closest verified catalogue candidate retaining both encoder and power-off brake is [23HS30-5004D-E1000-B280](https://www.omc-stepperonline.com/s-series-nema-23-closed-loop-stepper-motor-2-0nm-283-28oz-in-with-2-8nm-396-51oz-in-electromagnetic-brake-23hs30-5004d-e1000-b280), observed at **US$68.10** before shipping/taxes. It is a plausible cross-brand integration, not a supplier-confirmed plug-and-play replacement.

| Specification | CS-M22323B | 23HS30-5004D-E1000-B280 |
|---|---:|---:|
| Holding torque | 2.3 Nm | 2.0 Nm |
| Phase current | 5 A | 5 A |
| Phase resistance | 0.4 Ω | 0.42 Ω |
| Phase inductance | 1.7 mH | 1.7 mH |
| Motor | 2-phase, 1.8°, NEMA23 | 2-phase, 1.8°, NEMA23 |
| Shaft diameter / projection | 8 / 21 mm | 8 / 21 mm |
| Encoder | 1,000 lines, differential | 1,000 PPR / 4,000 CPR, differential A/B |
| Brake holding torque | 2 Nm | 2.8 Nm |
| Brake release | 24 V, 4.5 W | 24 V, 3.72 W / 0.155 A listed |
| Body length including rear accessories | approximately 137.5 mm from drawing | 136.5 mm listed |
| Weight | 1.65 kg | 1.61 kg |

The replacement's NEMA23 designation and shaft dimensions match the EG23-G20-D8 and MGL23-G10-D8 input requirements. Its exact pilot/bolt pattern drawing still needs verification; the PDF download was blocked. The product links a torque curve, but its image was not successfully inspected, so no running torque is asserted at 420/840 rpm. Similar resistance/inductance and a 13% lower holding torque make it a credible comparison candidate, not proof of equal dynamic performance.

[Leadshine CS2RS manual](https://www.leadshine.com/upfiles/downloads/19d35f99f92b99e0417f74134a985aff_1669716112793.pdf) confirms the D507's 20–50 VDC supply, 7 A peak current, incremental encoder support and +5 V encoder supply. Its CN6 mapping is VCC=5, GND=6, EB+=9, EB−=10, EA+=11, EA−=12; Z+/Z− at 7/8 are optional. The STEPPERONLINE motor page lists encoder pins VCC=2, GND=3, EA+=1, EA−=13, EB+=11, EB−=12. These are different connectors/pin assignments: make a signal-mapped harness, not an assumed straight-through cable. Check connector orientation and current production drawings before wiring. The motor listing's encoder logic-low voltage appears inconsistent (4.2–4.3 V); obtain a corrected electrical specification rather than relying on that field.

Driver Pr7.01 (register 0x0233) uses quadrature counts: **4,000 CPR** for this 1,000 PPR encoder. Configure phase current and tune the motor; D507's 7 A rating is peak, not RMS. Pr5.00 is peak current in 0.1 A units, and the manual advises approximately 1.2–1.4 times motor phase current. Confirm current convention with the motor supplier. Verify encoder polarity/phase relationship before loaded operation.

The CS2RS digital outputs are rated 100 mA at 24 V (50 mA recommended), so the 155 mA brake must use an external relay or suitably rated transistor stage with appropriate suppression. The manual supports configurable brake output (0x24) and release/engagement delays Pr4.19/Pr4.20. The original Leadshine brake also exceeds that direct-output current limit. Modbus commands remain a driver function; replacing a compatible bare motor does not require it to implement Modbus.

Recommendation: evaluate **23HS30-5004D-E1000-B280 + CS2RS-D507 + EG23-G20-D8**, verifying at least approximately 0.64 Nm motor running torque at 840 rpm plus acceleration margin for the provisional 12 Nm drum target. An MGL10 option instead needs approximately 0.89 Nm at 420 rpm for 8 Nm output and remains subject to its 10 Nm gearbox limit. The 1.2 Nm 23HS22-4004D-E1000-B280 is a smaller encoder/brake alternative but not the preferred headroom choice. Encoder-only motors and brake-only motors do not replace the complete CS-M22323B feature set.

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
