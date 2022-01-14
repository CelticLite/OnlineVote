'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
const { json } = require('stream/consumers');

class MyWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
        this.assets = new Set(); // AssetIDs of any assets created by this workload
        this.ballots = [];
    }
    
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);

        // Create Voter(s)
        var voterIDs = []
        for (let i=0; i<this.roundArguments.ballots; i++) {
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

        // Create ballots
        for (let i=0; i<this.roundArguments.ballots; i++) {
            const ballotId = `${voterIDs[i]}-${electionId}`;
            // Shuffle itemIds
            var j, x, a;
            for (a = itemIds.length - 1; a > 0; a--) {
                j = Math.floor(Math.random() * (a + 1));
                x = itemIds[a];
                itemIds[a] = itemIds[j];
                itemIds[j] = x;
            }

            console.log(`Worker ${this.workerIndex}: Creating ballot ${ballotId}`);
            const request = {
                contractId: this.roundArguments.contractId,
                contractFunction: 'NewBallot',
                invokerIdentity: this.roundArguments.invokerId,
                contractArguments: [voterIDs[i], electionId, JSON.stringify(itemIds)],
                readOnly: false
            };

            this.ballots.push(ballotId)
            await this.sutAdapter.sendRequests(request);
        }
        console.log("Created ballots: ", this.ballots)
    }
    
    async submitTransaction() {
        const voterId = `${this.workerIndex}_voter_${Math.floor(Math.random()*this.roundArguments.ballots)}`;
        const ballotId = `${voterId}-${this.workerIndex}_election`;
        
        const myArgs = {
            contractId: this.roundArguments.contractId,
            contractFunction: 'GetBallot',
            invokerIdentity: this.roundArguments.invokerId,
            contractArguments: [ballotId],
            readOnly: true
        };

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
