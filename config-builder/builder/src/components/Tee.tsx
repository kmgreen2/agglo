import * as React from "react"
import {Condition, ConditionBuilder, ConditionJSONProps, SetConditionFn, UndefinedCondition} from "./Condition";
import {Process} from "./Processes";
import {External} from "./Externals";
import {ChangeEvent, RefObject} from "react";
import {Transformer} from "./Transformer";
import {Form} from "react-bootstrap";
import {MainModal} from "./MainModal";

interface TeeBuilderProps {
    tee?: Tee
    setProcessState: (process: Process) => (void);
    validateProcessState: (process: Process, isEdit: boolean) => (void);
    getExternalState: () => Array<External>;
    getProcessState: () => Array<Process>;
}

interface TeeBuilderState {
    tee: Tee
}

function GetTransfomers(getProcessState: () => Array<Process>): Array<string> {
    return getProcessState().filter((process: Process) => {
        return (process instanceof Transformer);
    }).map((process: Process) => {
        return process.name;
    })
}

export class TeeBuilder extends React.Component<TeeBuilderProps, TeeBuilderState> {
    isEdit: boolean
    additionalBodyInputRef: RefObject<HTMLTextAreaElement>

    constructor(props: TeeBuilderProps) {
        super(props);
        let tee: Tee | undefined = this.props.tee;
        this.isEdit = true;
        this.additionalBodyInputRef = React.createRef<HTMLTextAreaElement>();

        if (this.props.tee == null) {
            this.isEdit = false;
            tee = new Tee({
                name: "",
                teeSpec: new TeeSpec(UndefinedCondition, "", "", ""),
            })
        }

        if (tee) {
            this.state = {
                tee: tee
            }
        }
    }

    componentDidMount() {
    }

    onChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        this.setState((prev: TeeBuilderState) => {
            let tee = prev.tee.copy();
            if (event.target.id === "teeName") {
                tee.name = event.target.value;
            } else if (event.target.id === "outputConnector") {
                tee.teeSpec.outputConnector = event.target.value;
            } else if (event.target.id === "transformerRef") {
                tee.teeSpec.transformerRef = event.target.value;
            }
            return {
                tee: tee
            }
        })
    }

    canAddCondition = (teeSpec: TeeSpec): boolean => {
        return teeSpec.outputConnector !== "";
    }

    saveConditionHandler = (): SetConditionFn => {
        return (condition: Condition) => {
            this.setState((prev: TeeBuilderState) => {
                let teeCopy = prev.tee.copy();
                teeCopy.teeSpec.condition = condition;
                return {
                    tee: teeCopy
                }
            })
        }
    }

    handleSave = (): any => {
        const errMessage = this.props.validateProcessState(this.state.tee, this.isEdit);
        if (errMessage == null) {
            if (this.additionalBodyInputRef.current) {
                let tee = this.state.tee.copy();
                tee.teeSpec.additionalBody = JSON.parse(this.additionalBodyInputRef.current.value);
                this.setState({
                    tee: tee
                })
            }
            this.props.setProcessState(this.state.tee);
        } else {
            return errMessage;
        }
        return null;
    }

    conditionForm = (teeSpec: TeeSpec): (JSX.Element) => {
        return <>
            <ConditionBuilder saveConditionHandler={this.saveConditionHandler()}
                              condition={teeSpec.condition}
                              enableButton={this.canAddCondition(teeSpec)} />
        </>
    }

    renderForm = (): (JSX.Element) => {
        const teeSpec = this.state.tee.teeSpec;
        const teeKey = this.state.tee.name;
        const defaultChoice = "Choose...";
        let defaultOutputConnector = defaultChoice;
        let defaultTransformerRef = defaultChoice;
        let conditionGroup = null;
        let additionalBody = teeSpec.additionalBody;

        if (teeSpec.additionalBody == {}) {
            additionalBody = {};
        }

        if (teeSpec.outputConnector && teeSpec.outputConnector !== "") {
            defaultOutputConnector = teeSpec.outputConnector;
        }

        if (teeSpec.transformerRef !== "") {
            defaultTransformerRef = teeSpec.transformerRef;
        }

        if (teeSpec.condition !== UndefinedCondition) {
            let condition = teeSpec.condition;
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
                <Form.Group controlId="teeName">
                    <Form.Label>Tee Name</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={teeKey}/>
                    <Form.Text className="text-muted"> Unique Tee name </Form.Text>
                </Form.Group>
                <Form.Group controlId="outputConnector">
                    <Form.Label>Output Connector</Form.Label>
                    <Form.Control as="select" onChange={this.onChange} defaultValue={defaultOutputConnector}>
                        <option key={defaultChoice}>{defaultChoice}</option>
                        {this.props.getExternalState().map(t => <option key={t.name}>{t.name}</option>)}
                    </Form.Control>
                    <Form.Text className="text-muted">
                        Choose an output connector
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId="additionalBody">
                    <Form.Label>Additional Body</Form.Label>
                    <Form.Control as="textarea" style={{ height: "400px" }}
                                  ref={this.additionalBodyInputRef}>
                        {JSON.stringify(additionalBody, null, 2)}
                    </Form.Control>
                    <Form.Text className="text-muted">
                        Add additional JSON body parameters to send
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId="transformerRef">
                    <Form.Label>Transformer</Form.Label>
                    <Form.Control as="select" onChange={this.onChange} defaultValue={defaultTransformerRef}>
                        <option key={defaultChoice}>{defaultChoice}</option>
                        {GetTransfomers(this.props.getProcessState).map(t => <option key={t}>{t}</option>)}
                    </Form.Control>
                    <Form.Text className="text-muted">
                        Choose an optional transformer
                    </Form.Text>
                </Form.Group>
                {conditionGroup}
                <Form.Group>
                    {this.conditionForm(teeSpec)}
                </Form.Group>
            </Form>
        </>
    }

    render() {
        let title = "Add Tee";
        if (this.isEdit) {
            title = "Update Tee";
        }
        return <>
            <MainModal title={title} renderForm={this.renderForm} handleSave={this.handleSave}/>
        </>
    }
}

export class TeeSpec {
    condition: Condition
    outputConnector: string
    transformerRef: string
    additionalBody: {}

    constructor(condition: Condition, outputConnector: string, transformerRef: string, additionalBody: {}) {
        this.outputConnector = outputConnector;
        this.transformerRef = transformerRef;

        if (additionalBody == null) {
            this.additionalBody = {}
        } else {
            this.additionalBody = additionalBody;
        }

        if (condition) {
            this.condition = condition;
        } else {
            this.condition = UndefinedCondition;
        }
    }

    copy = (): TeeSpec => {
        if (this.condition !== UndefinedCondition) {
            return new TeeSpec(this.condition.copy(), this.outputConnector, this.transformerRef, this.additionalBody);
        } else {
            return new TeeSpec(this.condition, this.outputConnector, this.transformerRef, this.additionalBody);
        }
    }
}

export interface TeeJSONProps {
    name: string
    condition: ConditionJSONProps
    outputConnectorRef: string
    transformerRef: string
    additionalBody: {}
}

interface TeeProps {
    name: string
    teeSpec: TeeSpec
}

interface TeeState {
}

export class Tee extends React.Component<TeeProps, TeeState> {
    name: string
    teeSpec: TeeSpec

    constructor(props: TeeProps) {
        super(props);
        this.name = props.name;
        this.teeSpec = props.teeSpec;
    }

    toString = ():string => {
        return JSON.stringify(this.toJSON());
    }

    static fromJSON = (props: TeeJSONProps): Tee => {
        return new Tee({
            name: props.name,
            teeSpec: new TeeSpec(
                Condition.fromJSON(props.condition),
                props.outputConnectorRef,
                props.transformerRef,
                props.additionalBody
            )
        })
    }

    toJSON = (): {} => {
        return {
            tee: {
                name: this.name,
                condition: this.teeSpec.condition.toJSON(),
                outputConnectorRef: this.teeSpec.outputConnector,
                transformerRef: this.teeSpec.transformerRef,
                additionalBody: this.teeSpec.additionalBody
            }
        }
    }

    copy = ():Tee => {
        return new Tee({
            name: this.name,
            teeSpec: this.teeSpec
        })
    }

    render() {
        return <>
            {this.name}
        </>
    }
}