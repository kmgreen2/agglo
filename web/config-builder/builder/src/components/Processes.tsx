import * as React from "react";
import {Annotator, AnnotatorBuilder, AnnotatorJSONProps} from "./Annotator";
import {ListGroup, ListGroupItem, Form, Button} from "react-bootstrap";
import {ChangeEvent} from "react";
import {Transformer, TransformerBuilder, TransformerJSONProps} from "./Transformer";
import {Completer, CompleterBuilder, CompleterJSONProps} from "./Completer";
import {External} from "./Externals";
import {Aggregator, AggregatorBuilder, AggregatorJSONProps} from "./Aggregator";
import {Tee, TeeBuilder, TeeJSONProps} from "./Tee";
import {Filter, FilterBuilder, FilterJSONProps} from "./Filter";
import {Spawner, SpawnerBuilder, SpawnerJSONProps} from "./Spawner";
import {CRC32} from "../common/Hash";
import {Continuation, ContinuationBuilder} from "./Continuation";
import {Entwine, EntwineBuilder} from "./Entwine";

export enum ProcessType {
    Annotator = "Annotator",
    Aggregator = "Aggregator",
    Completer = "Completer",
    Filter = "Filter",
    Spawner = "Spawner",
    Tee = "Tee",
    Transformer = "Transformer",
    Continuation = "Continuation",
    Entwine = "Entwine"
}

export function StrToProcessType(strProcessType: string): ProcessType {
    return ProcessType[strProcessType as keyof typeof ProcessType];
}

export function ProcessTypeToStr(operatorType: ProcessType): string {
    return operatorType
}

export function GetProcessTypes(): Array<string> {
    return Object.keys(ProcessType).filter((v) => v !== "Unknown")
}

export interface Process {
    name: string
    toJSON: () => {}
}

export type ProcessJSONProps = AggregatorJSONProps | AnnotatorJSONProps | CompleterJSONProps | FilterJSONProps |
    SpawnerJSONProps | TeeJSONProps | TransformerJSONProps;

export interface ProcessesProps {
    setProcessState: (process: Process)=>(void);
    deleteProcess: (process: Process)=>(void);
    validateProcessState: (process: Process, isEdit: boolean) => any;
    getProcessState: ()=> Array<Process>;
    getExternalState: ()=> Array<External>;
}

export interface ProcessesState {
    processSelector: string
}

export class Processes extends React.Component<ProcessesProps, ProcessesState> {
    builderId: number
    showModal: boolean
    showModalType: string
    constructor(props: ProcessesProps) {
        super(props);
        this.builderId = 0;
        this.showModal = false;
        this.showModalType = "";

        this.state = {
            processSelector: ProcessType.Annotator
        }
    }

    onProcessSelectChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>)  => {
        if (event.target.id === "processTypeSelector") {
            this.setState({
                processSelector: event.target.value
            })
        }
    }

    render() {
        let builder: JSX.Element = <></>
        if (this.state.processSelector === ProcessType.Annotator) {
            builder = <AnnotatorBuilder setProcessState={this.props.setProcessState}
                                        validateProcessState={this.props.validateProcessState}/>
        } else if (this.state.processSelector === ProcessType.Transformer) {
            builder = <TransformerBuilder setProcessState={this.props.setProcessState}
                                          validateProcessState={this.props.validateProcessState}/>
        } else if (this.state.processSelector === ProcessType.Completer) {
            builder = <CompleterBuilder setProcessState={this.props.setProcessState}
                                  validateProcessState={this.props.validateProcessState}
                                  getExternalState={this.props.getExternalState}/>
        } else if (this.state.processSelector === ProcessType.Aggregator) {
            builder = <AggregatorBuilder setProcessState={this.props.setProcessState}
                                        validateProcessState={this.props.validateProcessState}
                                        getExternalState={this.props.getExternalState}/>
        } else if (this.state.processSelector === ProcessType.Tee) {
            builder = <TeeBuilder setProcessState={this.props.setProcessState}
                                         validateProcessState={this.props.validateProcessState}
                                         getExternalState={this.props.getExternalState}
                                         getProcessState={this.props.getProcessState}/>
        } else if (this.state.processSelector === ProcessType.Filter) {
            builder = <FilterBuilder setProcessState={this.props.setProcessState}
                                  validateProcessState={this.props.validateProcessState} />
        } else if (this.state.processSelector === ProcessType.Spawner) {
            builder = <SpawnerBuilder setProcessState={this.props.setProcessState}
                                     validateProcessState={this.props.validateProcessState} />
        } else if (this.state.processSelector === ProcessType.Continuation) {
            builder = <ContinuationBuilder setProcessState={this.props.setProcessState}
                                      validateProcessState={this.props.validateProcessState} />
        } else if (this.state.processSelector === ProcessType.Entwine) {
            builder = <EntwineBuilder setProcessState={this.props.setProcessState}
                                      getExternalState={this.props.getExternalState}
                                      validateProcessState={this.props.validateProcessState} />
        }

return <>
    <ListGroup>
                {this.props.getProcessState().map((process) => {
                    if (process instanceof Annotator) {
                        return <>
                            <ListGroupItem>
                                <ListGroup horizontal={true}>
                                    <ListGroupItem>
                                        <Annotator key={process.name} name={process.name}
                                                   annotations={(process as Annotator).annotations}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <AnnotatorBuilder key={CRC32(process.toString())} setProcessState={this.props.setProcessState}
                                                          validateProcessState={this.props.validateProcessState}
                                                          annotator={process as Annotator}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <Button onClick={() => this.props.deleteProcess(process)}>Delete</Button>
                                    </ListGroupItem>
                                </ListGroup>
                            </ListGroupItem>
                        </>
                    } else if (process instanceof Transformer) {
                        return <>
                            <ListGroupItem>
                                <ListGroup horizontal={true}>
                                    <ListGroupItem>
                                        <Transformer key={process.name} name={process.name}
                                                     transformations={process.transformations}
                                                     forwardInputFields={process.forwardInputFields}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <TransformerBuilder key={CRC32(process.toString())} setProcessState={this.props.setProcessState}
                                                            validateProcessState={this.props.validateProcessState}
                                                            transformer={process}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <Button onClick={() => this.props.deleteProcess(process)}>Delete</Button>
                                    </ListGroupItem>
                                </ListGroup>
                            </ListGroupItem>
                        </>
                    } else if (process instanceof Completer) {
                        return <>
                            <ListGroupItem>
                                <ListGroup horizontal={true}>
                                    <ListGroupItem>
                                        <Completer key={process.name} name={process.name}
                                                     completion={process.completion}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <CompleterBuilder key={CRC32(process.toString())}
                                                          setProcessState={this.props.setProcessState}
                                                          validateProcessState={this.props.validateProcessState}
                                                          getExternalState={this.props.getExternalState}
                                                          completer={process}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <Button onClick={() => this.props.deleteProcess(process)}>Delete</Button>
                                    </ListGroupItem>
                                </ListGroup>
                            </ListGroupItem>
                        </>
                    } else if (process instanceof Aggregator) {
                        return <>
                            <ListGroupItem>
                                <ListGroup horizontal={true}>
                                    <ListGroupItem>
                                        <Aggregator key={process.name} name={process.name}
                                                   aggregation={process.aggregation}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <AggregatorBuilder key={CRC32(process.toString())}
                                                          setProcessState={this.props.setProcessState}
                                                          validateProcessState={this.props.validateProcessState}
                                                          getExternalState={this.props.getExternalState}
                                                          aggregator={process}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <Button onClick={() => this.props.deleteProcess(process)}>Delete</Button>
                                    </ListGroupItem>
                                </ListGroup>
                            </ListGroupItem>
                        </>
                    } else if (process instanceof Tee) {
                        return <>
                            <ListGroupItem>
                                <ListGroup horizontal={true}>
                                    <ListGroupItem>
                                        <Tee key={process.name} name={process.name}
                                                    teeSpec={process.teeSpec}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <TeeBuilder key={CRC32(process.toString())}
                                                    setProcessState={this.props.setProcessState}
                                                    validateProcessState={this.props.validateProcessState}
                                                    getExternalState={this.props.getExternalState}
                                                    getProcessState={this.props.getProcessState}
                                                    tee={process}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <Button onClick={() => this.props.deleteProcess(process)}>Delete</Button>
                                    </ListGroupItem>
                                </ListGroup>
                            </ListGroupItem>
                        </>
                    } else if (process instanceof Filter) {
                        return <>
                            <ListGroupItem>
                                <ListGroup horizontal={true}>
                                    <ListGroupItem>
                                        <Filter key={process.name} name={process.name}
                                             filterSpec={process.filterSpec}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <FilterBuilder key={CRC32(process.toString())}
                                                       setProcessState={this.props.setProcessState}
                                                    validateProcessState={this.props.validateProcessState}
                                                    filter={process}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <Button onClick={() => this.props.deleteProcess(process)}>Delete</Button>
                                    </ListGroupItem>
                                </ListGroup>
                            </ListGroupItem>
                        </>
                    } else if (process instanceof Spawner) {
                        return <>
                            <ListGroupItem>
                                <ListGroup horizontal={true}>
                                    <ListGroupItem>
                                        <Spawner key={process.name} name={process.name}
                                                spawnerSpec={process.spawnerSpec}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <SpawnerBuilder key={CRC32(process.toString())}
                                                        setProcessState={this.props.setProcessState}
                                                       validateProcessState={this.props.validateProcessState}
                                                       spawner={process}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <Button onClick={() => this.props.deleteProcess(process)}>Delete</Button>
                                    </ListGroupItem>
                                </ListGroup>
                            </ListGroupItem>
                        </>
                    } else if (process instanceof Continuation) {
                        return <>
                            <ListGroupItem>
                                <ListGroup horizontal={true}>
                                    <ListGroupItem>
                                        <Continuation key={process.name} name={process.name}
                                                 condition={process.condition} />
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <ContinuationBuilder key={CRC32(process.toString())}
                                                        setProcessState={this.props.setProcessState}
                                                        validateProcessState={this.props.validateProcessState}
                                                        continuation={process}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <Button onClick={() => this.props.deleteProcess(process)}>Delete</Button>
                                    </ListGroupItem>
                                </ListGroup>
                            </ListGroupItem>
                        </>
                    } else if (process instanceof Entwine) {
                        return <>
                            <ListGroupItem>
                                <ListGroup horizontal={true}>
                                    <ListGroupItem>
                                        <Continuation key={process.name} name={process.name}
                                                      condition={process.condition} />
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <EntwineBuilder key={CRC32(process.toString())}
                                                             setProcessState={this.props.setProcessState}
                                                             validateProcessState={this.props.validateProcessState}
                                                             getExternalState={this.props.getExternalState}
                                                             entwine={process}/>
                                    </ListGroupItem>
                                    <ListGroupItem>
                                        <Button onClick={() => this.props.deleteProcess(process)}>Delete</Button>
                                    </ListGroupItem>
                                </ListGroup>
                            </ListGroupItem>
                        </>
                    }
                    return <></>
                })}
                <ListGroupItem>
                    <ListGroup horizontal={true}>
                        <ListGroupItem>
                            <Form>
                            <Form.Group controlId="processTypeSelector">
                                <Form.Label>Process Type</Form.Label>
                                <Form.Control as="select" value={this.state.processSelector} onChange={this.onProcessSelectChange}>
                                    {GetProcessTypes().map(t => <option key={t}>{t}</option>)}
                                </Form.Control>
                            </Form.Group>
                            </Form>
                        </ListGroupItem>
                        <ListGroupItem>
                            {builder}
                        </ListGroupItem>
                    </ListGroup>
                </ListGroupItem>
            </ListGroup>
        </>
    }
}
