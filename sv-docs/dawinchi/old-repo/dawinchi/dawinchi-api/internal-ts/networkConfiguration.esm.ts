import fs from 'fs';
import JFile from 'jfile';
import { exec } from 'child_process';

interface NetworkObject {
    ip: string;
    gateway: string;
    dns: string;
}

class NetworkConfiguration {
    debug: boolean;
    networkDevice: string;

    constructor(){
        this.debug = true;
        this.networkDevice = 'eth0';
    }

    private async _setDHCP(state: boolean, networkDevice: string = this.networkDevice): Promise<string | Error> {
        let abort: boolean = false;
        if (typeof (networkDevice) === 'boolean') {
            state = networkDevice;
            networkDevice = this.networkDevice;
        };
        if (typeof (state) !== 'boolean') return Promise.reject('Please specify a boolean state!');
        const mode = await this._getDHCP(networkDevice);
        if (state === mode) abort = true;
        if (abort) return Promise.reject(`Interface ${networkDevice} is already set to ${state === true ? 'dhcp' : 'manual'} mode`);

        const interfaces = fs.readFileSync('/etc/network/interfaces').toString().split('\n');
        const line = new JFile('/etc/network/interfaces').grep(`iface  ${networkDevice}  inet`, true);

        interfaces.splice(line[0].i, 1, `iface  ${networkDevice}  inet  ${state === true ? 'dhcp' : 'manual'}`);

        const newFile = interfaces.join('\n');

        return new Promise((resolve, reject) => {
            fs.writeFile('/etc/network/interfaces', newFile, err => {
                if (err) return reject({ msg: 'There was an error saving the network configuration', err });
                resolve(state === true ? 'dhcp' : 'manual');
            });
        });
    }


    /**
     * Check the DHCP state of networkDevice interface. Defaults to this.networkDevice (eth0)
     * @param {string} networkDevice 
     * @returns 
     */
    private async _getDHCP(networkDevice: string = this.networkDevice): Promise<boolean | Error> {
        try {
            const interfaces = new JFile('/etc/network/interfaces');
            const line = interfaces.grep(`iface  ${networkDevice}  inet`);
            if (line.toString().includes('dhcp')) return Promise.resolve(true);
            return Promise.resolve(false);
        } catch (error) {
            return Promise.reject(error);
        }
    }

    /**
     * Set the static IP of the interface, defaults to this.networkDevice ('eth0')
     * @param {string} networkDevice Example: 'eth0'
     * @param {object} networkObject - Object containing the network settings to set for the interface
     * @example
     * this._setStaticInterface('eth0', {
     *      ip: '10.20.30.40',
     *      gateway: '10.20.30.1',
     *      dns: '1.1.1.1'
     * })
     */
    public async setStaticInterface(networkDevice: string = this.networkDevice, networkObject: NetworkObject): Promise<boolean | Error>{
        return new Promise((resolve, reject) => {
            if(typeof(networkDevice) === 'object'){ //Check if the first parameter is an object. If so, networkDevice has been omitted and should default to this.networkDevice.
                networkObject = networkDevice;      //Move the network object from networkDevice to networkObject
                networkDevice = this.networkDevice; //Assign the default network device.
            }

            const dhcpcd = new JFile('/etc/dhcpcd.conf');
            const line = dhcpcd.grep(`interface ${networkDevice}`, true); //Look for an already configured static interface matching networkDevice
            
            if(networkObject === undefined){
                if(line.length > 0){ //Array not empty, static configuration exists.
                    const newFile = dhcpcd.lines.join('\n').replace(/#/g, '');
                    fs.writeFile('/etc/dhcpcd.conf', newFile, err => {
                        if(err) return reject({msg: 'There was an error saving the network configuration', err}); //E.g access permissions error.
                        resolve(true); //Resolve before restarting the network interface to acknowledge the command. Restarting the network interface will take some time.
                        exec(`ifdown ${networkDevice} && ifup ${networkDevice}`);
                    });
                } else { //There's no static config for this interface
                    return reject('Please add a static configuration for this interface!');
                }
            }
            if(networkObject === undefined) return reject('Please supply a networkObject!');

            const {ip, gateway, dns} = networkObject; //Destructure networkObject.
            if(ip)      this._verifyIPAddress(ip).catch(reject); 
            if(gateway) this._verifyIPAddress(gateway).catch(reject);
            if(dns)     this._verifyIPAddress(dns).catch(reject);
            
            if(line.length === 0){
                if(!ip || !gateway || !dns) return reject(`Please supply IP, gateway and DNS since there are no static configuration for ${networkDevice}`);
            }

            let staticExists = false;
            if(line.length > 0) staticExists = true;                      //Array not empty, static configuration exists.
            const configuration = fs.readFileSync('/etc/dhcpcd.conf').toString().split('\n'); //Read interfaces file

            if(staticExists){
                for(let i = line[0].i; i < line[0].i + 5; i++){
                    let ipConfigured, gatewayConfigured, dnsConfigured;
                    console.log(dhcpcd.lines[i]);
                    //If 'interface' is present but networkDevice isn't, it's the wrong static config, so get out of dodge.
                    if(dhcpcd?.lines[i]?.includes('interface') && !dhcpcd.lines[i].includes(networkDevice)) break; 
                    if(dhcpcd?.lines[i]?.includes('interface')) continue; //Dont do anything to this line
                    if(dhcpcd?.lines[i] === undefined) break; //EOF

                    if(dhcpcd.lines[i].includes('ip_address') && ip)            configuration[i] = `static ip_address=${ip}/24`;
                    if(dhcpcd.lines[i].includes('routers') && gateway)          configuration[i] = `static routers=${gateway}`;
                    if(dhcpcd.lines[i].includes('domain_name_servers') && dns)  configuration[i] = `static domain_name_servers=${dns}`;
                
                    if(i === line[0].i + 5){
                        if(!ipConfigured && ip)             configuration.splice(line[0].i + 1, 0, `static ip_address=${ip}/24`);
                        if(!gatewayConfigured && gateway)   configuration.splice(line[0].i + 1, 0, `static routers=${gateway}`);
                        if(!dnsConfigured && dns)           configuration.splice(line[0].i + 1, 0, `static domain_name_servers=${dns}`);
                    }
                }
            }else{
                configuration[configuration.length]               = ``; //Gimme some space
                configuration[configuration.length]               = `interface ${networkDevice}`;
                if(ip) configuration[configuration.length]        = `static ip_address=${ip}/24`;
                if(gateway) configuration[configuration.length]   = `static routers=${gateway}`;
                if(dns) configuration[configuration.length]       = `static domain_name_servers=${dns}`;
            }
            console.log(configuration);

            const newFile = configuration.join('\n');
            fs.writeFile('/etc/dhcpcd.conf', newFile, err => {
                if(err) return reject({msg: 'There was an error saving the network configuration', err}); //E.g access permissions error.
                resolve(true); //Resolve before restarting the network interface to acknowledge the command. Restarting the network interface will take some time.
                exec(`ifdown ${networkDevice} && ifup ${networkDevice}`);
            });
        });
    }

    /**
     * Delete or disable a static configuration for networkDevice.
     * @param {string} networkDevice - Network device to configure.
     * @param {boolean} disable - True to disable configuration instead of deleting it. Disabling means it's possible to easily enable it later.
     * @returns 
     */
    private async _deleteStaticInterface(networkDevice: string = this.networkDevice, disable: boolean): Promise<boolean | Error>{
        return new Promise((resolve, reject) => {
            const dhcpcd = new JFile('/etc/dhcpcd.conf');
            const line = dhcpcd.grep(`interface ${networkDevice}`, true); //Find the static configuration if it exists. If not, line will be an empty array.
            if(line.length === 0) return reject('This interface is not statically configured'); //Empty array, static configuration doesn't exist.
            const configuration = fs.readFileSync('/etc/dhcpcd.conf').toString().split('\n'); //Read interfaces file

            if(disable){ //Don't delete the configuration, just disable it by commenting it out.
                for(let k = line[0].i; k < line[0].i + 5; k++){ //5 because there are only so many possible static parameters. Likely to be only 3
                    if(dhcpcd?.lines[k]?.includes(networkDevice) || dhcpcd?.lines[k]?.includes('static')) configuration[k] = `#${configuration[k]}`; //Remove the static entries for networkDevice
                    if(dhcpcd?.lines[k]?.includes('interface') && !dhcpcd?.lines[k]?.includes(networkDevice)) break; //Stop if you hit another static interface declaration!
                }
            }else{
                for(let k = line[0].i; k < line[0].i + 5; k++){ //5 because there are only so many possible static parameters. Likely to be only 3
                    if(dhcpcd?.lines[k]?.includes(networkDevice) || dhcpcd?.lines[k]?.includes('static')) configuration.splice(k, 1, ''); //Remove the static entries for networkDevice
                    if(dhcpcd?.lines[k]?.includes('interface') && !dhcpcd?.lines[k]?.includes(networkDevice)) break; //Stop if you hit another static interface declaration!
                }
            }

            const newFile = configuration.join('\n').trimEnd(); //Join array and create a string ready to write to file. Also, remove trailing newlines or they will accumulate over time.
            fs.writeFile('/etc/dhcpcd.conf', newFile, err => {
                if(err) return reject({msg: 'There was an error saving the network configuration', err}); //E.g access permissions error.
                resolve(true); //Resolve before restarting the network interface to acknowledge the command. Restarting the network interface will take some time.
                exec(`ifdown ${networkDevice} && ifup ${networkDevice}`);
            });
        });
    }

    /**
     * Verify that IP paramtere is a valid IP
     * @param {string} IP 
     * @returns 
     */
    private async _verifyIPAddress(IP: string): Promise<void | Error>{
        return new Promise((resolve, reject) => {
            //RegExp for checking a valid IP. AKA voodoo.
            const r = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/; 
            console.log(r.test(IP))
            if(r.test(IP)) return resolve();
            return reject();
        });
    }

    /**
     * Toggle DHCP either true or false.
     * @param {boolean} state 
     * @param {object} networkObject
     * @param {string} networkDevice 
     * @example
     * this.toggleDHCP(false, {
     *      ip: '10.20.30.40',
     *      gateway: '10.20.30.1',
     *      dns: '1.1.1.1'
     * }, 'eth0')
     */
    public async toggleDHCP(state: boolean, networkObject: NetworkObject, networkDevice: string = this.networkDevice): Promise<void | Error>{
        return new Promise((resolve, reject) => {
            if(typeof(state) !== 'boolean') return reject('Provide a boolean, true or false, to toggle DHCP!');
            if(state === true){
                this._setDHCP(true).then(() => {
                    this._deleteStaticInterface(undefined, true).then((res) => {
                        console.log('returned', res);
                        resolve();
                    }).catch(console.log);
                }).catch(console.log);
            }else{
                this._setDHCP(false).then(() => {
                    this.setStaticInterface(networkDevice, networkObject).catch((err) => { // setStaticInterface failed, set DHCP back to true.
                        console.log(err);
                        this._setDHCP(true).then(() => {}).catch(console.log);
                    });
                }).catch(console.log);
            }
        });
    }
}
export default NetworkConfiguration;

const network = new NetworkConfiguration();
// network.toggleDHCP(true, {
//    dns: "1.1.1.1"
// });

// network.toggleDHCP(true);
// network.toggleDHCP(true).catch(console.log);

// network._deleteStaticInterface();
