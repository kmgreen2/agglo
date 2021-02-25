import * as React from "react";
import {ChangeEvent} from "react";
import {Condition, ConditionBuilder, ConditionJSONProps, SetConditionFn, UndefinedCondition} from "./Condition";
import {Process} from "./Processes";
import {Form} from "react-bootstrap";
import {External, ExternalType} from "./Externals";
import {MainModal} from "./MainModal";
import {Aggregation, AggregationType} from "./Aggregator";

interface EntwineBuilderProps {
    entwine?: Entwine
    setProcessState: (process: Process) => (void);
    validateProcessState: (process: Process, isEdit: boolean) => (void);
    getExternalState: () => Array<External>;
}

interface EntwineBuilderState {
    entwine: Entwine
}

export class EntwineBuilder extends React.Component<EntwineBuilderProps, EntwineBuilderState> {
    isEdit: boolean

    constructor(props: EntwineBuilderProps) {
        super(props);
        let entwine: Entwine | undefined = this.props.entwine;
        this.isEdit = true;
        if (this.props.entwine == null) {
            entwine = new Entwine({
                name: "",
                streamStateStoreRef: "",
                objectStoreRef: "",
                pemPath: "",
                subStreamID: "",
                tickerEndpoint: "",
                tickerInterval: 0,
                condition: UndefinedCondition
            })
        }

        if (entwine) {
            this.state = {
                entwine: entwine
            }
        }
    }

    componentDidMount() {
    }

    onChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        this.setState((prev: EntwineBuilderState) => {
            let entwine = prev.entwine.copy();
            if (event.target.id === "entwineName") {
                entwine.name = event.target.value;
            } else if (event.target.id === "streamStateStore") {
                entwine.streamStateStoreRef = event.target.value;
            } else if (event.target.id === "objectStore") {
                entwine.objectStoreRef = event.target.value;
            } else if (event.target.id === "pemPath") {
                entwine.pemPath = event.target.value;
            } else if (event.target.id === "subStreamID") {
                entwine.subStreamID = event.target.value;
            } else if (event.target.id === "tickerEndpoint") {
                entwine.tickerEndpoint = event.target.value;
            } else if (event.target.id === "tickerInterval") {
                entwine.tickerInterval = parseInt(event.target.value);
            }
            return {
                entwine: entwine
            }
        })
    }

    handleSave = (): any => {
        const errMessage = this.props.validateProcessState(this.state.entwine, this.isEdit);
        if (errMessage == null) {
            this.props.setProcessState(this.state.entwine);
        } else {
            return errMessage;
        }
        return null;
    }

    canAddCondition = (entwine: Entwine): boolean => {
        return entwine.streamStateStoreRef !== "" &&
            entwine.name !== "" &&
            entwine.objectStoreRef !== "" &&
            entwine.pemPath !== "" &&
            entwine.subStreamID !== "" &&
            entwine.tickerEndpoint !== ""
    }

    saveConditionHandler = (): SetConditionFn => {
        return (condition: Condition) => {
            this.setState((prev) => {
                let entwine = prev.entwine.copy()
                entwine.condition = condition
                return {
                    entwine: entwine
                }
            })
        }
    }

    conditionForm = (entwine: Entwine): (JSX.Element) => {
        return <>
            <ConditionBuilder saveConditionHandler={this.saveConditionHandler()}
                              condition={entwine.condition}
                              enableButton={this.canAddCondition(entwine)} />
        </>
    }

    renderForm = (): (JSX.Element) => {
        const entwine = this.state.entwine;
        const defaultChoice = "Choose...";
        let defaultStateStore = defaultChoice;
        let defaultObjectStore = defaultChoice;
        let conditionGroup = null;

        if (entwine.streamStateStoreRef && entwine.streamStateStoreRef !== "") {
            defaultStateStore = entwine.streamStateStoreRef;
        }
        if (entwine.objectStoreRef && entwine.objectStoreRef !== "") {
            defaultObjectStore = entwine.objectStoreRef;
        }

        if (entwine.condition !== UndefinedCondition) {
            let condition = entwine.condition;
            conditionGroup = <Form.Group key={condition.toString()}>
                <Form.Label>Condition</Form.Label>
                <Form.Row>
                    <Condition operatorType={condition.operatorType}
                               operator={condition.operator}
                               lhs={condition.lhs}
                               rhs={condition.rhs} />
                </Form.Row>
            </Form.Group>
        }

        return <>
            <Form>
                <Form.Group controlId="entwineName">
                    <Form.Label>Entwine Name</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={entwine.name}/>
                    <Form.Text className="text-muted"> Unique Entwine name </Form.Text>
                </Form.Group>
                <Form.Group controlId="streamStateStore">
                    <Form.Label>Stream State Store</Form.Label>
                    <Form.Control as="select" onChange={this.onChange} defaultValue={defaultStateStore}>
                        <option key={defaultChoice}>{defaultChoice}</option>
                        {this.props.getExternalState().filter(t => t.externalType === ExternalType.KVStore).map(t => <option key={t.name}>{t.name}</option>)}
                    </Form.Control>
                    <Form.Text className="text-muted">
                        Choose a state store for the underlying stream
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId="objectStore">
                    <Form.Label>Object Store</Form.Label>
                    <Form.Control as="select" onChange={this.onChange} defaultValue={defaultObjectStore}>
                        <option key={defaultChoice}>{defaultChoice}</option>
                        {this.props.getExternalState().filter(t => t.externalType === ExternalType.ObjectStore).map(t => <option key={t.name}>{t.name}</option>)}
                    </Form.Control>
                    <Form.Text className="text-muted">
                        Choose a object store used to store blobs
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId="pemPath">
                    <Form.Label>PEM Path</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={entwine.pemPath}/>
                    <Form.Text className="text-muted"> Absolute path to PEM file </Form.Text>
                </Form.Group>
                <Form.Group controlId="subStreamID">
                    <Form.Label>SubStream ID</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={entwine.subStreamID}/>
                    <Form.Text className="text-muted"> ID used to identify the stream this process will write to (UUID suggested) </Form.Text>
                </Form.Group>
                <Form.Group controlId="tickerEndpoint">
                    <Form.Label>Ticker Endpoint</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={entwine.tickerEndpoint}/>
                    <Form.Text className="text-muted"> Endpoint of the ticker service to anchor with </Form.Text>
                </Form.Group>
                <Form.Group controlId="tickerInterval">
                    <Form.Label>Stream Anchor Interval</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={entwine.tickerInterval}/>
                    <Form.Text className="text-muted"> Processing interval (in messages) to anchor with ticker </Form.Text>
                </Form.Group>
                {conditionGroup}
                <Form.Group>
                    {this.conditionForm(entwine)}
                </Form.Group>
            </Form>
        </>
    }

    render() {
        let title = "Add Entwine";
        if (this.isEdit) {
            title = "Update Entwine";
        }
        return <>
            <MainModal title={title} renderForm={this.renderForm} handleSave={this.handleSave}/>
        </>
    }
}

interface EntwineJSONProps {
    name: string
    streamStateStore: string
    objectStore: string
    pemPath: string
    subStreamID: string
    tickerEndpoint: string
    tickerInterval: number
    condition: ConditionJSONProps
}

interface EntwineProps {
    name: string
    streamStateStoreRef: string
    objectStoreRef: string
    pemPath: string
    subStreamID: string
    tickerEndpoint: string
    tickerInterval: number
    condition?: Condition
}

interface EntwineState {
}

export class Entwine extends React.Component<EntwineProps, EntwineState> {
    name: string
    streamStateStoreRef: string
    objectStoreRef: string
    pemPath: string
    subStreamID: string
    tickerEndpoint: string
    tickerInterval: number
    condition: Condition

    constructor(props: EntwineProps) {
        super(props);
        this.name = props.name;
        this.streamStateStoreRef = props.streamStateStoreRef;
        this.objectStoreRef = props.objectStoreRef;
        this.pemPath = props.pemPath;
        this.subStreamID = props.subStreamID;
        this.tickerEndpoint = props.tickerEndpoint;
        this.tickerInterval = props.tickerInterval;
        if (props.condition) {
            this.condition = props.condition;
        } else {
            this.condition = UndefinedCondition;
        }
    }

    static fromJSON = (props: EntwineJSONProps): Entwine => {
        return new Entwine({
            name: props.name,
            streamStateStoreRef: props.streamStateStore,
            objectStoreRef: props.objectStore,
            pemPath: props.pemPath,
            subStreamID: props.subStreamID,
            tickerEndpoint: props.tickerEndpoint,
            tickerInterval: props.tickerInterval,
            condition: Condition.fromJSON(props.condition)
        })
    }

    toJSON = (): {} => {
        return {
            entwine: {
                name: this.name,
                streamStateStore: this.streamStateStoreRef,
                objectStore: this.objectStoreRef,
                pemPath: this.pemPath,
                subStreamID: this.subStreamID,
                tickerEndpoint: this.tickerEndpoint,
                tickerInterval: this.tickerInterval,
                condition: this.condition.toJSON()
            }
        }
    }

    copy = (): Entwine => {
        return new Entwine({
            name: this.name,
            streamStateStoreRef: this.streamStateStoreRef,
            objectStoreRef: this.objectStoreRef,
            pemPath: this.pemPath,
            subStreamID: this.subStreamID,
            tickerEndpoint: this.tickerEndpoint,
            tickerInterval: this.tickerInterval,
            condition: this.condition.copy()
        })
    }

    render() {
        return <>
            {this.name}
        </>
    }
}