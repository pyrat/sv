# Upgrade Orin Nano Super firmware before installing JetPack 7

## Goal

Upgrade older-than-36.x firmware using a microSD card, without an Ubuntu host PC. A Mac can prepare the cards and USB installer.

## Requirements

- Jetson Orin Nano / Orin Nano Super Developer Kit
- Reliable 19V power supply
- 64GB+ UHS-1 microSD card and reader
- DisplayPort monitor and USB keyboard, or serial console
- Internet connection on the Jetson
- Balena Etcher on your computer
- For JetPack 7: 16GB+ USB stick and target microSD or NVMe SSD

**Never interrupt power during a firmware update. Writing images erases the selected storage.**

## 1. Check existing firmware

Power on and press **Esc** at the NVIDIA splash screen. Inspect the UEFI firmware version.

- Older than 36.0: follow the steps below.
- 36.x: already ready for the JetPack 7 ISO installation path.

## 2. Prepare JetPack 5.1.3

Download the updated SD-card image from:

https://developer.nvidia.com/embedded/jetpack-sdk-513

Use this specific updated filename:

`JP513-orin-nano-sd-card-image_b29.zip`

Write it to the microSD card using Balena Etcher.

Insert the card into the Jetson, boot, complete Ubuntu setup, and connect to the internet.

## 3. Complete the first bootloader update

Check that the background service has scheduled the update:

```bash
sudo systemctl status nv-l4t-bootloader-config
```

NVIDIA says a completed scheduling run shows the service as inactive with a successful exit status.

Once confirmed:

```bash
sudo reboot
```

Watch the firmware update and allow it to finish. The Jetson should boot back into JetPack 5.1.3.

## 4. Install the QSPI updater

Inspect the firmware:

```bash
sudo nvbootctrl dump-slots-info
```

NVIDIA's example at this stage shows `Current version: 35.5.0`.

Install the updater and, after installation succeeds, reboot:

```bash
sudo apt update
sudo apt install nvidia-l4t-jetson-orin-nano-qspi-updater
sudo reboot
```

Allow the QSPI update to finish completely.

Afterward, the JetPack 5.1.3 card is no longer compatible with the updated firmware. Once the update is finished, power off and replace the boot media.

## 5. Boot JetPack 6.2.1

Download its SD-card image:

https://developer.nvidia.com/embedded/jetpack-sdk-621

Write it to a microSD card with Etcher. You can reuse the previous card, erasing its contents.

Boot the Jetson and complete initial setup.

Check whether another firmware update was scheduled:

```bash
sudo systemctl status nv-l4t-bootloader-config
```

If scheduled, reboot and allow the update to complete:

```bash
sudo reboot
```

## 6. Verify firmware

```bash
sudo nvbootctrl dump-slots-info
```

Alternatively, press **Esc** during startup and read the UEFI version.

With 36.x firmware, proceed to the JetPack 7.2.1 USB installer.

## 7. Install JetPack 7.2.1 using USB

1. Download the Jetson ISO from:
   https://developer.nvidia.com/embedded/jetpack/downloads
2. Use Etcher to write the ISO to a 16GB+ USB stick. Do not write the ISO to microSD.
3. Install your target NVMe SSD or microSD card in the Jetson.
4. Insert the USB installer.
5. Boot with **Esc → Boot Manager → select the USB drive**.
6. When prompted for the QSPI firmware update, press **Y within 30 seconds**. NVIDIA warns that missing this prompt can cause installation to fail later.
7. Allow both firmware-update passes and associated reboots to finish.
8. Select **Install Jetson ISO r39.2.1**.
9. Choose the correct target storage. Installation erases it.
10. Finish installation and follow the reboot/removal prompts.
11. Complete Ubuntu first-boot setup.
12. Optionally select **Power Mode → MAXN SUPER**.

The USB ISO method supports JetPack 7.2 and later on this kit, not JetPack 6. It installs Jetson Linux; consult NVIDIA's JetPack SDK setup guide for additional SDK components.

## Official NVIDIA documentation

- [Firmware upgrade](https://docs.nvidia.com/jetson/orin-nano-devkit/user-guide/latest/update_firmware.html)
- [JetPack 7 USB installation](https://docs.nvidia.com/jetson/orin-nano-devkit/user-guide/latest/quick_start.html)
- [JetPack SDK component setup](https://docs.nvidia.com/jetson/orin-nano-devkit/user-guide/latest/setup_jetpack.html)

## Progress checklist

- [ ] Checked original firmware version
- [ ] Prepared updated JetPack 5.1.3 card
- [ ] Completed first bootloader update
- [ ] Installed QSPI updater and completed update
- [ ] Booted JetPack 6.2.1
- [ ] Completed any final JetPack 6 firmware update
- [ ] Confirmed 36.x firmware
- [ ] Prepared JetPack 7.2.1 USB installer
- [ ] Installed JetPack 7.2.1
