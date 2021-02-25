import * as React from "react"
import {
    Condition,
    ConditionBuilder,
    ConditionJSONProps,
    OperatorType,
    SetConditionFn,
    UndefinedCondition
} from "./Condition";
import {Process} from "./Processes";
import {External} from "./Externals";
import {ChangeEvent} from "react";
import {Completion} from "./Completer";
import {MainModal} from "./MainModal";
import {Form} from "react-bootstrap";

export enum AggregationType {
    Unknown = "Unknown",
    Sum = "Sum",
    Max = "Max",
    Min = "Min",
    Avg = "Avg",
    Count = "Count",
    DiscreteHistogram = "DiscreteHistogram"
}

export function StrToAggregationType(strAggregationType: string): AggregationType {
    return AggregationType[strAggregationType as keyof typeof AggregationType];
}

export function AggregationTypeToStr(aggregationType: AggregationType): string {
    return aggregationType
}

export function GetAggregationTypes(): Array<string> {
    return Object.keys(AggregationType).filter((v) => v !== "Unknown")
}

interface AggregatorBuilderProps {
    aggregator?: Aggregator
    setProcessState: (process: Process) => (void);
    validateProcessState: (process: Process, isEdit: boolean) => (void);
    getExternalState: () => Array<External>;
}

interface AggregatorBuilderState {
    aggregator: Aggregator
}

export class AggregatorBuilder extends React.Component<AggregatorBuilderProps, AggregatorBuilderState> {
    isEdit: boolean
    constructor(props: AggregatorBuilderProps) {
        super(props);
        let aggregator: Aggregator | undefined = this.props.aggregator;
        this.isEdit = true;

        if (aggregator == null) {
            this.isEdit = false;
            aggregator = new Aggregator({
                name: "",
                aggregation: new Aggregation(UndefinedCondition, "", "", AggregationType.Unknown,
                    [], false, false)
            })
        }

        if (aggregator) {
            this.state = {
                aggregator: aggregator
            }
        }
    }

    componentDidMount() {
    }

    onChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        this.setState((prev: AggregatorBuilderState) => {
            let aggregator = prev.aggregator.copy();
            if (event.target.id === "aggregatorName") {
                aggregator.name = event.target.value;
            } else if (event.target.id === "stateStore") {
                aggregator.aggregation.stateStore = event.target.value;
            } else if (event.target.id === "aggregationKey") {
                aggregator.aggregation.aggregationKey = event.target.value;
            } else if (event.target.id === "aggregationType") {
                aggregator.aggregation.aggregationType = StrToAggregationType(event.target.value);
            } else if (event.target.id === "groupByKeys") {
                aggregator.aggregation.groupByKeys = event.target.value.split(",");
            } else if (event.target.id === "asyncCheckpoint") {
                aggregator.aggregation.asyncCheckpoint = event.target.value === "on";
            } else if (event.target.id === "forwardState") {
                aggregator.aggregation.forwardState = event.target.value === "on";
            }
            return {
                aggregator: aggregator
            }
        })
    }

    handleSave = (): any => {
        const errMessage = this.props.validateProcessState(this.state.aggregator, this.isEdit);
        if (errMessage == null) {
            this.props.setProcessState(this.state.aggregator);
        } else {
            return errMessage;
        }
        return null;
    }

    canAddCondition = (aggregation: Aggregation): boolean => {
        return aggregation.aggregationType !== AggregationType.Unknown &&
            aggregation.aggregationKey !== "" && aggregation.stateStore !== ""
    }

    saveConditionHandler = (): SetConditionFn => {
        return (condition: Condition) => {
            this.setState((prev) => {
                let aggregator = prev.aggregator.copy()
                aggregator.aggregation.condition = condition
                return {
                    aggregator: aggregator
                }
            })
        }
    }

    conditionForm = (aggregation: Aggregation): (JSX.Element) => {
        return <>
            <ConditionBuilder saveConditionHandler={this.saveConditionHandler()}
                              condition={aggregation.condition}
                              enableButton={this.canAddCondition(aggregation)} />
        </>
    }

    renderForm = (): (JSX.Element) => {
        const aggregation = this.state.aggregator.aggregation;
        const aggregatorKey = this.state.aggregator.name;
        const defaultChoice = "Choose...";
        let defaultStateStore = defaultChoice;
        let defaultAggregationType = defaultChoice;
        let conditionGroup = null;

        if (aggregation.stateStore && aggregation.stateStore !== "") {
            defaultStateStore = aggregation.stateStore;
        }

        if (aggregation.aggregationType && aggregation.aggregationType !== AggregationType.Unknown) {
           defaultAggregationType = aggregation.aggregationType;
        }

        if (aggregation.condition !== UndefinedCondition) {
            let condition = aggregation.condition;
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
                <Form.Group controlId="aggregatorName">
                   <Form.Label>Aggregation Name</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={aggregatorKey}/>
                    <Form.Text className="text-muted"> Unique Aggregator name </Form.Text>
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
                <Form.Group controlId="aggregationType">
                    <Form.Label>AggregationType</Form.Label>
                    <Form.Control as="select" onChange={this.onChange} defaultValue={defaultAggregationType}>
                        <option key={defaultChoice}>{defaultChoice}</option>
                        {GetAggregationTypes().map(t => <option key={t}>{t}</option>)}
                    </Form.Control>
                    <Form.Text className="text-muted">
                        Choose a state store
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId="aggregationKey">
                    <Form.Label>Aggregation Key</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={aggregation.aggregationKey}/>
                    <Form.Text className="text-muted"> Aggregation key </Form.Text>
                </Form.Group>
                <Form.Group controlId="groupByKeys">
                    <Form.Label>Group By Keys</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={aggregation.groupByKeys.join(",")}/>
                    <Form.Text className="text-muted"> Choose the group by keys </Form.Text>
                </Form.Group>
                <Form.Group controlId="forwardState">
                    <Form.Label>Forward State</Form.Label>
                    <Form.Check
                        custom
                        type="checkbox"
                        onChange={this.onChange}
                        defaultChecked={aggregation.forwardState}
                    />
                    <Form.Text className="text-muted">Check to forward aggregation state to other processes</Form.Text>
                </Form.Group>
                <Form.Group controlId="asyncCheckpoint">
                    <Form.Label>Async Checkpoint</Form.Label>
                    <Form.Check
                        custom
                        type="checkbox"
                        defaultChecked={aggregation.asyncCheckpoint}
                        onChange={this.onChange}
                    />
                    <Form.Text className="text-muted">Check to checkpoint to external store asynchronously</Form.Text>
                </Form.Group>
                {conditionGroup}
                <Form.Group>
                    {this.conditionForm(aggregation)}
                </Form.Group>
            </Form>
        </>
    }

    render() {
        let title = "Add Aggregator";
        if (this.isEdit) {
            title = "Update Aggregator";
        }
        return <>
            <MainModal title={title} renderForm={this.renderForm} handleSave={this.handleSave}/>
        </>
    }
}

export class Aggregation {
    condition: Condition
    stateStore: string
    aggregationKey: string
    aggregationType: AggregationType
    groupByKeys: Array<string>
    asyncCheckpoint: boolean
    forwardState: boolean

    constructor(condition: Condition, stateStore: string, aggregationKey: string, aggregationType: AggregationType,
                groupByKeys: Array<string>, asyncCheckpoint: boolean, forwardState: boolean) {
        this.condition = condition;
        this.stateStore = stateStore;
        this.aggregationKey = aggregationKey;
        this.aggregationType = aggregationType;
        this.groupByKeys = groupByKeys;
        this.asyncCheckpoint = asyncCheckpoint;
        this.forwardState = forwardState;
    }

    copy = ():Aggregation => {
        if (this.condition !== UndefinedCondition) {
            return new Aggregation(
                this.condition.copy(),
                this.stateStore,
                this.aggregationKey,
                this.aggregationType,
                this.groupByKeys,
                this.asyncCheckpoint,
                this.forwardState
            )
        } else {
            return new Aggregation(
                this.condition,
                this.stateStore,
                this.aggregationKey,
                this.aggregationType,
                this.groupByKeys,
                this.asyncCheckpoint,
                this.forwardState
            )
        }
    }
}

export interface AggregatorJSONProps {
    name: string
    condition: ConditionJSONProps
    stateStore: string
    asyncCheckpoint: boolean
    forwardState: boolean
    aggregation: {
        key: string
        aggregationType: string
        groupByKeys: Array<string>
    }
}

interface AggregatorProps {
    name: string
    aggregation: Aggregation
}

interface AggregatorState {
}

export class Aggregator extends React.Component<AggregatorProps, AggregatorState> {
    name: string
    aggregation: Aggregation

    constructor(props: AggregatorProps) {
        super(props);
        this.name = props.name;
        this.aggregation = props.aggregation;
    }

    toString = ():string => {
        return JSON.stringify(this.toJSON());
    }

    static fromJSON = (props: AggregatorJSONProps): Aggregator => {
        return new Aggregator({
            name: props.name,
            aggregation: new Aggregation(
                Condition.fromJSON(props.condition),
                props.stateStore,
                props.aggregation.key,
                StrToAggregationType(props.aggregation.aggregationType),
                props.aggregation.groupByKeys,
                props.asyncCheckpoint,
                props.forwardState
            )
        })
    }

    toJSON = (): {} => {
        return {
            aggregator: {
                name: this.name,
                condition: this.aggregation.condition.toJSON(),
                stateStore: this.aggregation.stateStore,
                asyncCheckpoint: this.aggregation.asyncCheckpoint,
                forwardState: this.aggregation.forwardState,
                aggregation: {
                    key: this.aggregation.aggregationKey,
                    aggregationType: this.aggregation.aggregationType,
                    groupByKeys: this.aggregation.groupByKeys
                }
            }
        }
    }

    copy = (): Aggregator => {
        return new Aggregator({
            name: this.name,
            aggregation: this.aggregation
        })
    }

    render() {
        return <>
            {this.name}
        </>
    }
}
