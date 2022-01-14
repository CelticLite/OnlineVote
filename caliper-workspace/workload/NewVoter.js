'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
const { json } = require('stream/consumers');

class MyWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
        this.assets = new Set(); // AssetIDs of any assets created by this workload
    }
    
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);
    }
    
    async submitTransaction() {
        const voterId = `${this.workerIndex}_Voter${this.txIndex}`
        
        const myArgs = {
            contractId: this.roundArguments.contractId,
            contractFunction: 'NewVoter',
            invokerIdentity: this.roundArguments.invokerId,
            contractArguments: [voterId, "Name"],
            readOnly: false
        };

        this.assets.add(voterId);

        this.txIndex++;
        await this.sutAdapter.sendRequests(myArgs);
    }
    
    async cleanupWorkloadModule() {
        for (const element of this.assets.values()) {
            console.log(`Worker ${this.workerIndex}: Deleting asset ${element}`);
            const request = {
                contractId: this.roundArguments.contractId,
                contractFunction: 'DeleteAsset',
                invokerIdentity: this.roundArguments.invokerId,
                contractArguments: [element],
                readOnly: false
            };

            await this.sutAdapter.sendRequests(request);
        }
    }
}

function createWorkloadModule() {
    return new MyWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
