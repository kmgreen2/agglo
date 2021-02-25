import * as React from "react";
import {External, ExternalsProps} from "./Externals";
import {Process} from "./Processes";
import {ChangeEvent} from "react";
import { Map } from 'immutable'
import {v4 as uuidv4} from 'uuid';
import {
    Accordion,
    Button,
    ButtonGroup,
    ButtonToolbar,
    Card,
    Form,
    ListGroup,
    ListGroupItem,
    Tab,
    Tabs
} from "react-bootstrap";
import {MainModal} from "./MainModal";
import {CRC32} from "../common/Hash";

function GetProcesses(getProcessState: () => Array<Process>): Array<string> {
    return getProcessState().map((process: Process) => {
        return process.name;
    })
}

interface PipelineBuilderProps {
    pipeline?: Pipeline
    getExternalState: () => Array<External>;
    getProcessState: () => Array<Process>;
    validatePipelineState: (pipeline: Pipeline, isEdit: boolean) => any;
    setPipelineState: (pipeline: Pipeline) => (void);
}

interface PipelineBuilderState {
    pipeline: Pipeline
    processTab: string
}

export class PipelineBuilder extends React.Component<PipelineBuilderProps, PipelineBuilderState> {
    isEdit: boolean
    constructor(props: PipelineBuilderProps)  {
        super(props);
        let pipeline: Pipeline;

        this.isEdit = true;

        if (props.pipeline == null) {
            this.isEdit = false;
            pipeline = new Pipeline({
                name: "",
                processes: Array<PipelineProcess>(),
                checkPointConnectorRef: "",
                enableMetrics: false,
                enableTracing: false
            })
        } else {
            pipeline = props.pipeline;
        }

        this.state = {
            pipeline: pipeline,
            processTab: pipeline.name + "0"
        }
    }

    componentDidMount() {
    }

    onChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        let targetIdAry = event.target.id.split(":");
        let targetIdField = targetIdAry[0];

        this.setState((prev: PipelineBuilderState) => {
            let pipeline = prev.pipeline.copy();
            if (targetIdAry.length === 1) {
                if (targetIdField === "pipelineName") {
                    pipeline.name = event.target.value;
                } else if (targetIdField === "checkpointConnector") {
                    pipeline.checkPointConnectorRef = event.target.value;
                } else if (targetIdField === "pipelineEnableMetrics") {
                    pipeline.enableMetrics = !pipeline.enableMetrics;
                } else if (targetIdField === "pipelineEnableTracing") {
                    pipeline.enableTracing = !pipeline.enableTracing;
                }
            } else if (targetIdAry.length === 2) {
                let targetIdIdx = parseInt(targetIdAry[1]);
                let process = pipeline.processes[targetIdIdx].copy();
                if (targetIdField === "processRef")  {
                    process.processRef = event.target.value;
                } else if (targetIdField === "numRetries")  {
                    process.numRetries = parseInt(event.target.value);
                } else if (targetIdField === "initialBackOffMs")  {
                    process.initialBackOffMs = parseInt(event.target.value);
                } else if (targetIdField === "enableLatency")  {
                    process.enableLatency = !process.enableLatency;
                } else if (targetIdField === "enableCounter")  {
                    process.enableCounter = !process.enableCounter;
                } else if (targetIdField === "enableTracing")  {
                    process.enableTracing = !process.enableTracing;
                }
                pipeline.processes[targetIdIdx] = process;
            }
            return {
                pipeline: pipeline
            }
        })
    }

    handleSave = (): any => {
        const errMessage = this.props.validatePipelineState(this.state.pipeline, this.isEdit);
        if (errMessage == null) {
            this.props.setPipelineState(this.state.pipeline);
        } else {
            return errMessage;
        }
        return null;
    }

    setProcessTab = (k: string|null) => {
        if (k) {
            this.setState({
                processTab: k
            })
        }
    }

    deleteProcess = (processIndex: number) => {
        this.setState((prev: PipelineBuilderState) => {
            let pipeline = prev.pipeline.copy()
            let processes = pipeline.processes.filter((elm: PipelineProcess, idx: number) => {
                return idx !== processIndex;
            })
            pipeline.processes = processes;
            return {
                pipeline: pipeline
            }
        }, () =>
            this.setProcessTab(this.state.pipeline.name + `${this.state.pipeline.processes.length-1}`))
    }

    appendProcess = () => {
        this.setState((prev: PipelineBuilderState) => {
            let pipeline = prev.pipeline.copy();
            let newProcess = new PipelineProcess("", 0, 0,
                false, false, false);
            pipeline.processes = pipeline.processes.concat(newProcess);
            return {
                pipeline: pipeline
            }
        }, () =>
            this.setProcessTab(this.state.pipeline.name + `${this.state.pipeline.processes.length-1}`))
    }

    pipelineProcessForm = (pipelineProcess: PipelineProcess, processIndex: number): (JSX.Element) => {
        const defaultChoice = "Choose...";
        let defaultProcess = defaultChoice;
        if (pipelineProcess.processRef) {
            defaultProcess = pipelineProcess.processRef;
        }

        return <>
            <Form>
                <Form.Group controlId={`processRef:${processIndex}`}>
                    <Form.Label>Process</Form.Label>
                    <Form.Control as="select" onChange={this.onChange} defaultValue={defaultProcess}>
                        <option key={defaultChoice}>{defaultChoice}</option>
                        {GetProcesses(this.props.getProcessState).map(t => <option key={t}>{t}</option>)}
                    </Form.Control>
                    <Form.Text className="text-muted">
                        Choose a process
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId={`numRetries:${processIndex}`}>
                    <Form.Label>Num Retries</Form.Label>
                    <Form.Control type="text" onChange={this.onChange}
                                  defaultValue={pipelineProcess.numRetries}/>
                    <Form.Text className="text-muted">
                        Optional num retries for process
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId={`initialBackOffMs:${processIndex}`}>
                    <Form.Label>Retry Initial Backoff (ms)</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} disabled={pipelineProcess.numRetries === 0}
                                  defaultValue={pipelineProcess.initialBackOffMs}/>
                    <Form.Text className="text-muted">
                        Optional retry backoff in ms
                    </Form.Text>
                    <Form.Group controlId={`enableTracing:${processIndex}`}>
                        <Form.Label>Enable Tracing</Form.Label>
                        <Form.Check
                            custom
                            type="checkbox"
                            defaultChecked={pipelineProcess.enableTracing}
                            onChange={this.onChange}
                        />
                        <Form.Text className="text-muted">Enable tracing for this process</Form.Text>
                    </Form.Group>
                    <Form.Group controlId={`enableLatency:${processIndex}`}>
                        <Form.Label>Enable Latency Metrics</Form.Label>
                        <Form.Check
                            custom
                            type="checkbox"
                            defaultChecked={pipelineProcess.enableLatency}
                            onChange={this.onChange}
                        />
                        <Form.Text className="text-muted">Enable latency metrics for this process</Form.Text>
                    </Form.Group>
                    <Form.Group controlId={`enableCounter:${processIndex}`}>
                        <Form.Label>Enable Counter Metrics</Form.Label>
                        <Form.Check
                            custom
                            type="checkbox"
                            defaultChecked={pipelineProcess.enableCounter}
                            onChange={this.onChange}
                        />
                        <Form.Text className="text-muted">Enable counter metrics for this process</Form.Text>
                    </Form.Group>
                    <Form.Group>
                        <Button onClick={() => this.deleteProcess(processIndex)}>Delete</Button>
                    </Form.Group>
                </Form.Group>
            </Form>
        </>
    }

    moveProcessUp = (index: number) => {
        this.setState((prev: PipelineBuilderState) => {
            let pipeline = prev.pipeline.copy();
            const tmp = pipeline.processes[index];
            pipeline.processes[index] = pipeline.processes[index-1]
            pipeline.processes[index-1] = tmp;
            return {
                pipeline: pipeline,
                processTab: ""
            }
        })
    }

    moveProcessDown = (index: number) => {
        this.setState((prev: PipelineBuilderState) => {
            let pipeline = prev.pipeline.copy();
            const tmp = pipeline.processes[index];
            pipeline.processes[index] = pipeline.processes[index+1]
            pipeline.processes[index+1] = tmp;
            return {
                pipeline: pipeline,
                processTab: ""
            }
        })
    }

    renderForm = (): (JSX.Element) => {
        const processes = this.state.pipeline.processes;
        const pipelineKey = this.state.pipeline.name;
        let processIndex = 0;
        const processTab = this.state.processTab;
        const defaultChoice = "Choose...";
        let defaultOutputConnector = defaultChoice;

        if (this.state.pipeline.checkPointConnectorRef) {
            defaultOutputConnector = this.state.pipeline.checkPointConnectorRef;
        }

        return <>
            <Form>
                <Form.Group controlId="pipelineName">
                    <Form.Label>Pipeline Name</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={pipelineKey}/>
                    <Form.Text className="text-muted"> Unique Pipeline name </Form.Text>
                </Form.Group>
                <Accordion id="pipeline-tabs" onSelect={this.setProcessTab}>
                    {processes.map(process => {
                        let title = "New Process";
                        const myIndex = processIndex;
                        if (process.processRef) {
                            title = `${process.processRef}`;
                        }
                        const tab = <Card>
                            <Card.Header>
                                <Accordion.Toggle eventKey={pipelineKey + `${processIndex}`} as={Button} variant="link">
                                    {title}
                                </Accordion.Toggle>
                                <ButtonToolbar aria-label="process-up-down-button-toolbar">
                                    <ButtonGroup className="mr-2" aria-label="process-up-down-button-group">
                                        <Button  onClick={() => this.moveProcessUp(myIndex)} hidden={processIndex === 0}>Up</Button>
                                        <Button onClick={() => this.moveProcessDown(myIndex)} hidden={processIndex === processes.length-1}>Down</Button>
                                    </ButtonGroup>
                                </ButtonToolbar>
                            </Card.Header>
                            <Accordion.Collapse eventKey={pipelineKey + `${processIndex}`}>
                                <Card.Body>
                                    {this.pipelineProcessForm(process, processIndex)}
                                </Card.Body>
                            </Accordion.Collapse>
                        </Card>
                        processIndex++;
                        return tab;
                    })}
                </Accordion>
                <Button onClick={this.appendProcess}>Add Process</Button>
                <Form.Group controlId="checkpointConnector">
                    <Form.Label>Checkpoint Output Connector</Form.Label>
                    <Form.Control as="select" onChange={this.onChange} defaultValue={defaultOutputConnector}>
                        <option key={defaultChoice}>{defaultChoice}</option>
                        {this.props.getExternalState().map(t => <option key={t.name}>{t.name}</option>)}
                    </Form.Control>
                    <Form.Text className="text-muted">
                        Choose an output connector for storing checkpoints
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId={"pipelineEnableTracing"}>
                    <Form.Label>Enable Tracing</Form.Label>
                    <Form.Check
                        custom
                        type="checkbox"
                        defaultChecked={this.state.pipeline.enableTracing}
                        onChange={this.onChange}
                    />
                    <Form.Text className="text-muted">Enable tracing</Form.Text>
                </Form.Group>
                <Form.Group controlId={"pipelineEnableMetrics"}>
                    <Form.Label>Enable Metrics</Form.Label>
                    <Form.Check
                        custom
                        type="checkbox"
                        defaultChecked={this.state.pipeline.enableMetrics}
                        onChange={this.onChange}
                    />
                    <Form.Text className="text-muted">Enable metrics</Form.Text>
                </Form.Group>
            </Form>
        </>
    }

    render() {
        let title = "Add Pipeline";
        if (this.isEdit) {
            title = "Update pipeline";
        }
        return <>
            <MainModal title={title} renderForm={this.renderForm} handleSave={this.handleSave}/>
        </>
    }
}

interface PipelineProcessJSONProps {
    name: string
    retryStrategy: {
        numRetries: number,
        initialBackOffMs: number
    }
    instrumentation: {
        enableTracing: boolean,
        latency: boolean,
        counter: boolean
    }
}

class PipelineProcess {
    processRef:  string
    numRetries: number
    initialBackOffMs: number
    enableTracing: boolean
    enableLatency: boolean
    enableCounter: boolean

    constructor(processRef: string, numRetries: number, initialBackOffMs:number, enableTracing: boolean,
                enableLatency: boolean, enableCounter: boolean) {
        this.processRef = processRef;
        this.numRetries = numRetries;
        this.initialBackOffMs = initialBackOffMs;
        this.enableTracing = enableTracing;
        this.enableLatency = enableLatency;
        this.enableCounter = enableCounter;
    }

    toJSON = ():{} => {
        return {
            name: this.processRef,
            retryStrategy: {
                numRetries: this.numRetries,
                initialBackOffMs: this.initialBackOffMs
            },
            instrumentation: {
                enableTracing: this.enableTracing,
                latency: this.enableLatency,
                counter: this.enableCounter
            }
        }
    }

    static fromJSON = (props: PipelineProcessJSONProps): PipelineProcess => {
        let numRetries = 0;
        let initialBackOffMs = 0;
        let enableTracing = false;
        let enableLatency = false;
        let enableCounter= false;

        if (props.retryStrategy) {
            numRetries = props.retryStrategy.numRetries;
            initialBackOffMs = props.retryStrategy.initialBackOffMs;
        }

        if (props.instrumentation) {
            enableTracing = props.instrumentation.enableTracing;
            enableLatency = props.instrumentation.latency;
            enableCounter = props.instrumentation.counter;
        }

        return new PipelineProcess(props.name, numRetries, initialBackOffMs,
            enableTracing, enableLatency, enableCounter);
    }

    copy = (): PipelineProcess => {
        return new PipelineProcess(
            this.processRef,
            this.numRetries,
            this.initialBackOffMs,
            this.enableTracing,
            this.enableLatency,
            this.enableCounter
        )
    }
}

interface PipelineProps {
    name: string
    processes: Array<PipelineProcess>
    checkPointConnectorRef: string
    enableTracing: boolean
    enableMetrics: boolean
}

export interface PipelineJSONProps {
    name: string
    processes: Array<PipelineProcessJSONProps>
    checkpoint: {
        outputConnectorRef: string
    }
    enableTracing: boolean
    enableMetrics: boolean
}

interface PipelineState {
}

export class Pipeline extends React.Component<PipelineProps, PipelineState> {
    name: string
    processes: Array<PipelineProcess>
    checkPointConnectorRef: string
    enableTracing: boolean
    enableMetrics: boolean

    constructor(props: PipelineProps) {
        super(props);
        this.name = props.name;
        this.processes = props.processes;
        this.checkPointConnectorRef = props.checkPointConnectorRef;
        this.enableMetrics = props.enableMetrics;
        this.enableTracing = props.enableTracing;
    }

    toString = ():string => {
        return JSON.stringify(this.toJSON());
    }

    toJSON = ():{} => {
        if (this.checkPointConnectorRef.length > 0) {
            return {
                name: this.name,
                processes: this.processes.map(proc => proc.toJSON()),
                checkpoint: {
                    outputConnectorRef: this.checkPointConnectorRef
                },
                enableTracing: this.enableTracing,
                enableMetrics: this.enableMetrics
            }
        }
        return {
            name: this.name,
            processes: this.processes.map(proc => proc.toJSON()),
            enableTracing: this.enableTracing,
            enableMetrics: this.enableMetrics
        }
    }

    static fromJSON = (props: PipelineJSONProps) => {
        if (props.checkpoint === undefined || props.checkpoint == null) {
            props.checkpoint = {
                outputConnectorRef: ""
            }
        }
        return new Pipeline({
            name: props.name,
            processes: props.processes.map(procProps => PipelineProcess.fromJSON(procProps)),
            checkPointConnectorRef: props.checkpoint.outputConnectorRef,
            enableTracing: props.enableTracing,
            enableMetrics: props.enableMetrics
        })
    }

    copy = ():Pipeline => {
        return new Pipeline({
            name: this.name,
            processes: this.processes.map((process) => process.copy()),
            checkPointConnectorRef: this.checkPointConnectorRef,
            enableTracing: this.enableTracing,
            enableMetrics: this.enableMetrics,
        })
    }

    render() {
        return <>
            {this.name}
            </>
    }
}

export interface PipelinesProps {
    getMetadataState: () => Map<string, string>;
    setMetadataState: (key: string, value: string) => (void);
    getExternalState: () => Array<External>;
    getProcessState: () => Array<Process>;
    getPipelineState: () => Array<Pipeline>;
    validatePipelineState: (pipeline: Pipeline, isEdit: boolean) => any;
    setPipelineState: (pipeline: Pipeline) => (void);
    deletePipeline: (pipeline: Pipeline) => (void);
}

interface PipelinesState {
    partitionUuid: string
}

export class Pipelines extends React.Component<PipelinesProps, PipelinesState> {
    // Builders *need* a new instance each time, so we generate a
    //new key for each one
    builderId: number

    constructor(props: PipelinesProps) {
        super(props);
        this.builderId = 0;
        const metadata = props.getMetadataState()
        let partitionUuid: undefined| string = metadata.get("partitionUuid");

        if (partitionUuid) {
            this.state = {
                partitionUuid: partitionUuid
            }
        } else {
            this.state = {
                partitionUuid: uuidv4()
            }
        }
    }

    onChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        this.setState((prev: PipelinesState) => {
            let partitionUuid = "";
            if (event.target.id === "partitionUuid") {
                partitionUuid = event.target.value;
            }
            return {
                partitionUuid: partitionUuid
            }
        })
    }

    handleSave = (): any => {
        this.props.setMetadataState("partitionUuid", this.state.partitionUuid);
        return;
    }

    render() {
        let partitionUuid = this.state.partitionUuid;
        return <>
            <Form>
                <Form.Group controlId="partitionUuid">
                    <Form.Label>Partition UUID</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={partitionUuid}/>
                    <Form.Text className="text-muted"> Unique Continuation name </Form.Text>
                </Form.Group>
                <Form.Group controlId="saveUUID">
                    <Button variant="primary" onClick={this.handleSave}>
                        Save UUID
                    </Button>
                </Form.Group>
            </Form>
            <ListGroup>
                {this.props.getPipelineState().map((pipeline: Pipeline) =>
                    <ListGroupItem>
                        <ListGroup horizontal={true}>
                            <ListGroupItem>
                                <Pipeline key={pipeline.name} name={pipeline.name}  processes={pipeline.processes}
                                          checkPointConnectorRef={pipeline.checkPointConnectorRef}
                                          enableMetrics={pipeline.enableMetrics} enableTracing={pipeline.enableTracing}/>
                            </ListGroupItem>
                            <ListGroupItem>
                                <PipelineBuilder key={CRC32(pipeline.toString())} setPipelineState={this.props.setPipelineState}
                                                 validatePipelineState={this.props.validatePipelineState}
                                                 pipeline={pipeline} getExternalState={this.props.getExternalState}
                                                 getProcessState={this.props.getProcessState}/>
                            </ListGroupItem>
                            <ListGroupItem>
                                <Button onClick={() => this.props.deletePipeline(pipeline)}>Delete</Button>
                            </ListGroupItem>
                        </ListGroup>
                    </ListGroupItem>
                )}
            </ListGroup>
            <PipelineBuilder key={this.builderId++} setPipelineState={this.props.setPipelineState}
                             validatePipelineState={this.props.validatePipelineState}
                             getExternalState={this.props.getExternalState} getProcessState={this.props.getProcessState}/>
        </>
    }
}