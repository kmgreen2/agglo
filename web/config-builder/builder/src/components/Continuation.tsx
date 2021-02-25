import * as React from "react"
import {Condition, ConditionBuilder, ConditionJSONProps, SetConditionFn, UndefinedCondition} from "./Condition";
import {Process} from "./Processes";
import {External} from "./Externals";
import {Tee, TeeSpec} from "./Tee";
import {MainModal} from "./MainModal";
import {Form} from "react-bootstrap";
import {ChangeEvent} from "react";

interface ContinuationBuilderProps {
    continuation?: Continuation
    setProcessState: (process: Process) => (void);
    validateProcessState: (process: Process, isEdit: boolean) => (void);
}

interface ContinuationBuilderState {
    continuation: Continuation
}

export class ContinuationBuilder extends React.Component<ContinuationBuilderProps, ContinuationBuilderState> {
    isEdit: boolean

    constructor(props: ContinuationBuilderProps) {
        super(props);

        let continuation: Continuation | undefined = this.props.continuation;
        this.isEdit = true;

        if (this.props.continuation == null) {
            this.isEdit = false;
            continuation = new Continuation({
                name: "",
                condition: UndefinedCondition
            })
        }

        if (continuation) {
            this.state = {
                continuation: continuation
            }
        }
    }

    componentDidMount() {
    }

    onChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        this.setState((prev: ContinuationBuilderState) => {
            let continuation = prev.continuation.copy();
            if (event.target.id === "continuationName") {
                continuation.name = event.target.value;
            }
            return {
                continuation: continuation
            }
        })
    }

    canAddCondition = (continuation: Continuation): boolean => {
        return continuation.name !== "";
    }

    saveConditionHandler = (): SetConditionFn => {
        return (condition: Condition) => {
            this.setState((prev: ContinuationBuilderState) => {
                let continuationCopy = prev.continuation.copy();
                continuationCopy.condition = condition;
                return {
                    continuation: continuationCopy
                }
            })
        }
    }

    handleSave = (): any => {
        const errMessage = this.props.validateProcessState(this.state.continuation, this.isEdit);
        if (errMessage == null) {
            this.props.setProcessState(this.state.continuation);
        } else {
            return errMessage;
        }
        return null;
    }

    conditionForm = (continuation: Continuation): (JSX.Element) => {
        return <>
            <ConditionBuilder saveConditionHandler={this.saveConditionHandler()}
                              condition={continuation.condition}
                              enableButton={this.canAddCondition(continuation)} />
        </>
    }

    renderForm = (): (JSX.Element) => {
        let continuation = this.state.continuation;
        let continuationName = continuation.name;
        let conditionGroup = null;

        if (continuation.condition !== UndefinedCondition) {
            let condition = continuation.condition;
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

        return <Form>
            <Form.Group controlId="continuationName">
                <Form.Label>Continuation Name</Form.Label>
                <Form.Control type="text" onChange={this.onChange} defaultValue={continuationName}/>
                <Form.Text className="text-muted"> Unique Continuation name </Form.Text>
            </Form.Group>
            {conditionGroup}
            <Form.Group>
                {this.conditionForm(continuation)}
            </Form.Group>
        </Form>
    }

    render() {
        let title = "Add Continuation";
        if (this.isEdit) {
            title = "Update Continuation";
        }
        return <>
            <MainModal title={title} renderForm={this.renderForm} handleSave={this.handleSave}/>
        </>
    }
}

interface ContinuationJSONProps {
    name: string
    condition: ConditionJSONProps
}

interface ContinuationProps {
    name: string
    condition: Condition
}

interface ContinuationState {
}

export class Continuation extends React.Component<ContinuationProps, ContinuationState> {
    name: string
    condition: Condition

    constructor(props: ContinuationProps) {
        super(props);
        this.name = props.name;
        this.condition = props.condition;
    }

    toString = ():string => {
        return JSON.stringify(this.toJSON());
    }

    static fromJSON = (props: ContinuationJSONProps): Continuation => {
        return new Continuation({
            name: props.name,
            condition: Condition.fromJSON(props.condition)
        })
    }

    toJSON = (): {} => {
        return {
            continuation: {
                name: this.name,
                condition: this.condition.toJSON(),
            }
        }
    }

    copy = ():Continuation => {
        return new Continuation({
            name: this.name,
            condition: this.condition.copy()
        })
    }

    render() {
        return <>
            {this.name}
        </>
    }
}
