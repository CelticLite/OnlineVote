'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
const { json } = require('stream/consumers');

class MyWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
        this.items = [];
        this.assets = new Set();
    }
    
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);
        // Create Voter(s)
        var voterIDs = []
        for (let i=0; i<this.roundArguments.tx; i++) {
            const assetID = `${this.workerIndex}_voter_${i}`;
            console.log(`Worker ${this.workerIndex}: Creating voter ${assetID}`);
            const request = {
                contractId: this.roundArguments.contractId,
                contractFunction: 'NewVoter',
                invokerIdentity: this.roundArguments.invokerId,
                contractArguments: [assetID,`Voter${i}`],
                readOnly: false
            };
            voterIDs.push(assetID)
            this.assets.add(assetID)
            await this.sutAdapter.sendRequests(request);
        }

        console.log("Created voters: ", voterIDs)

        // Create VotableItem(s)
        var itemIds = []
        for (let i=0; i<this.roundArguments.items; i++) {
            const assetID = `${this.workerIndex}_vitem_${i}`;
            console.log(`Worker ${this.workerIndex}: Creating vitem ${assetID}`);
            const request = {
                contractId: this.roundArguments.contractId,
                contractFunction: 'NewVotableItem',
                invokerIdentity: this.roundArguments.invokerId,
                contractArguments: [assetID,`Canidate${i}`,'Decription'],
                readOnly: false
            };
            itemIds.push(assetID)
            this.assets.add(assetID)
            await this.sutAdapter.sendRequests(request);
        }

        this.items = itemIds;

        // Create election
        const electionId = `${this.workerIndex}_election`;
        console.log(`Worker ${this.workerIndex}: Creating election ${electionId}`);
        const request = {
            contractId: this.roundArguments.contractId,
            contractFunction: 'NewElection',
            invokerIdentity: this.roundArguments.invokerId,
            contractArguments: [electionId,'Election', JSON.stringify(voterIDs), JSON.stringify(itemIds)],
            readOnly: false
        };
        this.assets.add(electionId)
        await this.sutAdapter.sendRequests(request);
    }
    
    async submitTransaction() {
        const itemIds = this.items;

        // Shuffle itemIds
        var j, x, i;
        for (i = itemIds.length - 1; i > 0; i--) {
            j = Math.floor(Math.random() * (i + 1));
            x = itemIds[i];
            itemIds[i] = itemIds[j];
            itemIds[j] = x;
        }
            
        
        const myArgs = {
            contractId: this.roundArguments.contractId,
            contractFunction: 'CastVote',
            invokerIdentity: this.roundArguments.invokerId,
            contractArguments: [`${this.workerIndex}_voter_${this.txIndex}`, `${this.workerIndex}_election`, JSON.stringify(itemIds)],
            readOnly: false
        };

        // Add ballot id to the asset list
        this.assets.add(`${this.workerIndex}_voter_${this.txIndex}-${this.workerIndex}_election`)

        //console.log(`Worker ${this.workerIndex}: Submitting tx ${this.workerIndex}_voter_${voterId}-${this.workerIndex}_election`);

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
