# Estimate of Work - Smartwinch Phase 1

## 1. Scope

The scope of this document is to provide an estimate of the work required to complete the project. The estimate is based on the requirements and the architecture of the system.

Requirements described here: [Requirements](https://wiki.imenco-dev.com/projects/daWinchi/SRS) on the Imenco Wiki follow the MoSCoW method. We have included all Must have requirements in this estimate and refer to the deliverable as __Phase 1__.

## 2. Estimate

The bottom line estimate is __55 working days__ until completion for 2 engineers of which both are senior engineers, one of which is also a senior software architect. The majority of the API has already been designed and implemented. The majority of the work remaining concerns implementation of hardware functions and integration of the hardware with the API layer.

Over the course of the development period one of the engineers will be working on the project full time, while the other will be working on the project part time. The part time engineer will be working on the project 50% of the time. In practice he will work 100% for a period of time and complete the majority of the remaining API work.

The hardware and hardware API interface will be implemented by the full time engineer and is best suited for this engineer as he has the most experience with hardware and embedded systems. This work is sequential in nature and will not benefit from parallelization.

This translates to a delivery date of __30th of November 2023__.


## 3. Work Breakdown Structure

The work breakdown structure is based on the architecture of the system. The system is divided into 2 layers: API and Hardware. The hardware layer is divided into 2 sub layers: Hardware Interface and API Interface.

### Hardware Interface

* Exploratory phase button interface
* Remote lock from buttons
* Fully featured interface to ADS1115
* Interface to LS7366R plus counter HW
* Remote lock on button panel
* Hardware reboot functionality
* Factory reset functionality
* Precision mode via button panel
* Automatic control - x, z coords
* Algorithm for speed calculation
* Fault condition indicate on Button / LED panel
* Locally operate the winch up/down/in/out
* Service mode implementation (stop remote and local control of the winch)
* Set Home position functionality via button panel
* Go to Home position functionality via button panel

Total: 40 man days

### Hardware API Interface

* Status endpoint
* Remote operation of winch up/down/in/out
* Set speeds for both directions individually
* General integration between hardware "firmware" and API

Total: 15 man days

### Software API (already partly implemented)

* Remote operation of winch up/down/in/out
* Logging functionality
* Logging API
* Factory configuration API
* Automatic control API with scheduling

Total: 20 man days

### Project Management

* Project management of the project internally at Createview

Total: 5 days

## Testing

* Test plan
* Test cases
* Test execution
* Automated test suite

Total: 5 days

### Total

75 man days development
5 days testing
5 days project management

Some of the work can be completed in parallel so the end to end time for delivery is 60 working days.

This estimate starts from the 7 of September 2023 and results in a delivery date of 30th of November 2023.

## 4. Risks

The risks associated with this project are:

* Unforeseen hardware issues - this is mitigated by the exploratory phase some of which has already been completed.
* Unforeseen software issues - this is mitigated by the exploratory phase some of which has already been completed.
* Physical testing overhead - we will need some humans to eventually test the system in the test environment which requires some resource at Imenco to help with this. We aim to mitigate this by having a test plan and test cases ready for the test team to execute. We also plan to have a prototype with buttons and LEDs that we can use in our own lab environment to test portions of the hardware.






