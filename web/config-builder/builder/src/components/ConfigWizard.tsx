import React from 'react';
import {External, Externals} from './Externals'
import {Tab, Tabs} from 'react-bootstrap'
import { Map } from 'immutable'
import {CRC32} from '../common/Hash'
import {ConditionBuilder} from "./Condition";
import {Process, Processes} from "./Processes";
import {Pipeline, Pipelines} from "./Pipelines";
import {Editor} from "./Editor";

interface WizardProps{}
interface WizardState{
    metadata: Map<string, string>
    externals: Map<string, External>
    processes: Map<string, Process>
    pipelines: Map<string, Pipeline>
}

export class ConfigWizard extends React.Component<WizardProps, WizardState> {
    renderCounter: number
    constructor(props: WizardProps) {
        super(props);
        this.renderCounter = 0;
        this.state = {
            metadata: Map<string, string>(),
            externals: Map<string, External>(),
            processes: Map<string, Process>(),
            pipelines: Map<string, Pipeline>()
        }
    }

    setMetadataState = (key: string, value: string) => {
        this.setState(state => {
            const metadata = state.metadata.set(key, value);
            return {
                metadata,
            }
        })
    }

    getMetadataState = ():Map<string, string> => {
        return this.state.metadata;
    }

    setProcessState = (process: Process) => {
        this.setState(state => {
            const processes = state.processes.set(process.name, process);
            return {
                processes,
            }
        });
    }

    deleteProcess = (process: Process) => {
        this.setState(state => {
            const processes = state.processes.delete(process.name)
            return {
                processes,
            }
        });
    }

    validateProcessState = (process: Process, isEdit: boolean): any => {
        // Edits will not change the name, so no need to check for name conflicts
        if (!isEdit) {
            const validName = Array.from(this.state.processes.keys()).every((k: string) => {
                return process.name !== k;
            })

            if (!validName) {
                return "invalid name: " + process.name;
            }
        }
        return null;
    }

    getProcessState = ():Array<Process> => {
        return Array.from(this.state.processes.values())
    }

    setExternalState = (external: External) => {
        this.setState(state => {
            const externals = state.externals.set(external.name, external);
            return {
                externals,
            }
        });
    }

    deleteExternal = (external: External) => {
        this.setState(state => {
            const externals = state.externals.delete(external.name)
            return {
                externals,
            }
        });
    }

    validateExternalState = (external: External, isEdit: boolean): any => {
        // Edits will not change the name, so no need to check for name conflicts
        if (!isEdit) {
            const validName = Array.from(this.state.externals.keys()).every((k: string) => {
                return external.name !== k;
            })

            if (!validName) {
                return "invalid name: " + external.name;
            }
        }
        return null;
    }

    getExternalState = ():Array<External> => {
        return Array.from(this.state.externals.values())
    }

    setPipelineState = (pipeline: Pipeline) => {
        this.setState(state => {
            const pipelines = state.pipelines.set(pipeline.name, pipeline);
            return {
                pipelines,
            }
        });
    }

    deletePipeline = (pipeline: Pipeline) => {
        this.setState(state => {
            const pipelines = state.pipelines.delete(pipeline.name)
            return {
                pipelines,
            }
        });
    }

    validatePipelineState = (pipeline: Pipeline, isEdit: boolean): any => {
        // Edits will not change the name, so no need to check for name conflicts
        if (!isEdit) {
            const validName = Array.from(this.state.pipelines.keys()).every((k: string) => {
                return pipeline.name !== k;
            })

            if (!validName) {
                return "invalid name: " + pipeline.name;
            }
        }
        return null;
    }

    getPipelineState = ():Array<Pipeline> => {
        return Array.from(this.state.pipelines.values())
    }

    render() {
        const externalsChanges = CRC32(Array.from(this.state.externals.values()).map(v => v.toString()).join(""));
        const processesChanges = CRC32(Array.from(this.state.processes.values()).map(v => v.toString()).join(""));
        const pipelinesChanges = CRC32(Array.from(this.state.pipelines.values()).map(v => v.toString()).join(""));

        this.renderCounter++;
        return <>
            <Tabs defaultActiveKey="externals" id="wizard-tabs">
               <Tab eventKey="externals" title="Externals">
                  <Externals key={externalsChanges} setExternalState={this.setExternalState} getExternalState={this.getExternalState}
                             validateExternalState={this.validateExternalState} deleteExternal={this.deleteExternal}/>
               </Tab>
                <Tab eventKey="processes" title="Processes">
                    <Processes key={processesChanges} setProcessState={this.setProcessState}
                               getProcessState={this.getProcessState} validateProcessState={this.validateProcessState}
                               deleteProcess={this.deleteProcess}
                               getExternalState={this.getExternalState}/>
                </Tab>
                <Tab eventKey="pipelines" title="Pipelines">
                    <Pipelines key={pipelinesChanges}
                               setMetadataState={this.setMetadataState}
                               getMetadataState={this.getMetadataState}
                               getProcessState={this.getProcessState}
                               validatePipelineState={this.validatePipelineState}
                               deletePipeline={this.deletePipeline}
                               setPipelineState={this.setPipelineState}
                               getPipelineState={this.getPipelineState}
                               getExternalState={this.getExternalState}/>
                </Tab>
                <Tab eventKey="editor" title="Editor">
                    <Editor key={this.renderCounter}
                            setMetadataState={this.setMetadataState}
                            getMetadataState={this.getMetadataState}
                            validatePipelineState={this.validatePipelineState}
                            deletePipeline={this.deletePipeline}
                            setPipelineState={this.setPipelineState}
                            getPipelineState={this.getPipelineState}
                            setExternalState={this.setExternalState} getExternalState={this.getExternalState}
                            validateExternalState={this.validateExternalState} deleteExternal={this.deleteExternal}
                            setProcessState={this.setProcessState}
                            getProcessState={this.getProcessState} validateProcessState={this.validateProcessState}
                            deleteProcess={this.deleteProcess}
                    />
                </Tab>
            </Tabs>
       </>
    }
}

