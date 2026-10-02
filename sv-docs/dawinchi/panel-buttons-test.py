import RPi.GPIO as GPIO
import time

# Set up GPIO mode and configure pins
GPIO.setmode(GPIO.BCM)
pins = [6, 14, 15, 16, 17]
GPIO.setup(pins, GPIO.IN, pull_up_down=GPIO.PUD_UP)

try:
    while True:
        pin_states = {}
        for pin in pins:
            pin_states[pin] = GPIO.input(pin)

        for pin, state in pin_states.items():
            print(f"GPIO {pin}: state={state}")

        print("-------------\n")

        time.sleep(0.5)

except KeyboardInterrupt:
    GPIO.cleanup()
