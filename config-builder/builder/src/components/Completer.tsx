import * as React from "react"
import {Condition, ConditionBuilder, ConditionJSONProps, SetConditionFn, UndefinedCondition} from "./Condition";
import {Process} from "./Processes";
import {ChangeEvent} from "react";
import {MainModal} from "./MainModal";
import {Form} from "react-bootstrap";
import {External} from "./Externals";

interface CompleterBuilderProps {
    completer?: Completer
    setProcessState: (process: Process) => (void);
    validateProcessState: (process: Process, isEdit: boolean) => (void);
    getExternalState: () => Array<External>;
}

interface CompleterBuilderState {
    completer: Completer
}

export class CompleterBuilder extends React.Component<CompleterBuilderProps, CompleterBuilderState> {
    isEdit: boolean

    constructor(props: CompleterBuilderProps) {
        super(props);
        let completer: Completer | undefined = this.props.completer;
        this.isEdit = true;

        if (this.props.completer == null) {
            this.isEdit = false;
            completer = new Completer({
                name: "",
                completion: new Completion(UndefinedCondition, "", [], 0)
            })
        }

        if (completer) {
            this.state = {
                completer: completer
            }
        }
    }

    componentDidMount() {
    }

    onChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        this.setState((prev: CompleterBuilderState) => {
            let completer = prev.completer.copy();
            if (event.target.id === "completerName") {
                completer.name = event.target.value;
            } else if (event.target.id === "stateStore") {
                completer.completion.stateStore = event.target.value;
            } else if (event.target.id === "joinKeys") {
                completer.completion.joinKeys = event.target.value.split(",");
            } else if (event.target.id === "timeoutMs") {
                completer.completion.timeoutMs = parseInt(event.target.value);
            }
            return {
                completer: completer
            }
        })
    }

    handleSave = (): any => {
        const errMessage = this.props.validateProcessState(this.state.completer, this.isEdit);
        if (errMessage == null) {
            this.props.setProcessState(this.state.completer);
        } else {
            return errMessage;
        }
        return null;
    }

    canAddCondition = (completion: Completion): boolean => {
        return completion.joinKeys.length > 0 && completion.stateStore !== ""
    }

    saveConditionHandler = (): SetConditionFn => {
        return (condition: Condition) => {
            this.setState((prev) => {
                let completer = prev.completer.copy()
                completer.completion.condition = condition
                return {
                    completer: completer
                }
            })
        }
    }

    conditionForm = (completion: Completion): (JSX.Element) => {
        return <>
            <ConditionBuilder saveConditionHandler={this.saveConditionHandler()}
                              condition={completion.condition}
                              enableButton={this.canAddCondition(completion)} />
        </>
    }

    renderForm = (): (JSX.Element) => {
        const completion = this.state.completer.completion;
        const completerKey = this.state.completer.name;
        const defaultChoice = "Choose...";
        let defaultStateStore = defaultChoice;
        let conditionGroup = null;

        if (completion.stateStore && completion.stateStore !== "") {
            defaultStateStore = completion.stateStore;
        }

        if (completion.condition !== UndefinedCondition) {
            let condition = completion.condition;
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
                <Form.Group controlId="completerName">
                    <Form.Label>Completer Name</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={completerKey}/>
                    <Form.Text className="text-muted"> Unique Completer name </Form.Text>
                </Form.Group>
                <Form.Group controlId="stateStore">
                    <Form.Label>State Store</Form.Label>
                    <Form.Control as="select" onChange={this.onChange} defaultValue={defaultStateStore}>
                        <option key={defaultChoice}>{defaultChoice}</option>
                        {this.props.getExternalState().map(t => <option key={t.name}>{t.name}</option>)}
                    </Form.Control>
                    <Form.Text className="text-muted">
                        Choose a state store
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId="joinKeys">
                    <Form.Label>Join Keys</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={completion.joinKeys.join(",")}/>
                    <Form.Text className="text-muted"> Choose the join keys </Form.Text>
                </Form.Group>
                <Form.Group controlId="timeoutMs">
                    <Form.Label>Timeout (ms)</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={completion.timeoutMs}/>
                    <Form.Text className="text-muted"> Specify a timeout in milliseconds</Form.Text>
                </Form.Group>
                {conditionGroup}
                <Form.Group>
                    {this.conditionForm(completion)}
                </Form.Group>
            </Form>
        </>
    }

    render() {
        let title = "Add Completer";
        if (this.isEdit) {
            title = "Update Completer";
        }
        return <>
            <MainModal title={title} renderForm={this.renderForm} handleSave={this.handleSave}/>
        </>
    }
}

export class Completion {
    condition: Condition
    stateStore: string
    joinKeys: Array<string>
    timeoutMs: number

    constructor(condition: Condition, stateStore: string, joinKeys: Array<string>, timeoutMs: number) {
        this.condition = condition;
        this.stateStore = stateStore;
        this.joinKeys = joinKeys;
        this.timeoutMs = timeoutMs;
    }

    copy = ():Completion => {
        if (this.condition !== UndefinedCondition) {
            return new Completion(
                this.condition.copy(),
                this.stateStore,
                Array.from(this.joinKeys),
                this.timeoutMs
            )
        } else {
            return new Completion(
                this.condition,
                this.stateStore,
                Array.from(this.joinKeys),
                this.timeoutMs
            )
        }
    }
}

export interface CompleterJSONProps {
    name: string
    condition: ConditionJSONProps
    stateStore: string
    completion: {
        joinKeys: Array<string>
        timeoutMs: number
    }
}

interface CompleterProps {
    name: string
    completion: Completion
}

interface CompleterState {
}

export class Completer extends React.Component<CompleterProps, CompleterState> {
    name: string
    completion: Completion

    constructor(props: CompleterProps) {
        super(props);
        this.name = props.name;
        this.completion = props.completion;
    }

    static fromJSON = (props: CompleterJSONProps): Completer => {
        return new Completer({
            name: props.name,
            completion: new Completion(Condition.fromJSON(props.condition), props.stateStore,
                props.completion.joinKeys, props.completion.timeoutMs)
        })
    }

    toString = ():string => {
        return JSON.stringify(this.toJSON());
    }

    toJSON = (): {} => {
        return {
            completer: {
                name: this.name,
                condition: this.completion.condition.toJSON(),
                stateStore: this.completion.stateStore,
                completion: {
                    joinKeys: this.completion.joinKeys,
                    timeoutMs: this.completion.timeoutMs
                }
            }
        }
    }

    copy = ():Completer => {
        return new Completer({
            name: this.name,
            completion: this.completion.copy()
        })
    }

    render() {
        return <>
            {this.name}
        </>
    }
}