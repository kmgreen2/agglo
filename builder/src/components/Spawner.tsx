import {Condition, ConditionBuilder, ConditionJSONProps, SetConditionFn, UndefinedCondition} from "./Condition";
import * as React from "react";
import {Process} from "./Processes";
import {ChangeEvent} from "react";
import {TeeSpec} from "./Tee";
import {Form} from "react-bootstrap";
import {MainModal} from "./MainModal";
import {spawn} from "child_process";

interface SpawnerBuilderProps {
    spawner?: Spawner
    setProcessState: (process: Process) => (void);
    validateProcessState: (process: Process, isEdit: boolean) => (void);
}

interface SpawnerBuilderState {
    spawner: Spawner
}

export class SpawnerBuilder extends React.Component<SpawnerBuilderProps, SpawnerBuilderState> {
    isEdit: boolean
    constructor(props: SpawnerBuilderProps) {
        super(props);
        let spawner: Spawner | undefined = this.props.spawner;
        this.isEdit = true;

        if (this.props.spawner == null) {
            this.isEdit = false;
            spawner = new Spawner({
                name: "",
                spawnerSpec: new SpawnerSpec(UndefinedCondition, 0, false, "", [])
            })
        }

        if (spawner) {
            this.state = {
                spawner: spawner
            }
        }
    }

    componentDidMount() {
    }

    onChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        this.setState((prev: SpawnerBuilderState) => {
            let spawner = prev.spawner.copy();
            if (event.target.id === "spawnerName") {
               spawner.name = event.target.value;
            } else if (event.target.id === "pathToExec") {
                spawner.spawnerSpec.pathToExec = event.target.value;
            } else if (event.target.id === "doSync") {
                spawner.spawnerSpec.doSync = event.target.value === "on";
            } else if (event.target.id === "delayInMs") {
                spawner.spawnerSpec.delayInMs = parseInt(event.target.value);
            } else if (event.target.id === "commandArgs") {
                spawner.spawnerSpec.commandArgs = event.target.value.split(",");
            }

            return {
                spawner: spawner
            }
        })
    }

    canAddCondition = (spawnerSpec: SpawnerSpec): boolean => {
        return spawnerSpec.pathToExec !== ""
    }

    saveConditionHandler = (): SetConditionFn => {
        return (condition: Condition) => {
            this.setState((prev: SpawnerBuilderState) => {
                let spawnerCopy = prev.spawner.copy();
                spawnerCopy.spawnerSpec.condition = condition;
                return {
                    spawner: spawnerCopy
                }
            })
        }
    }

    handleSave = (): any => {
        const errMessage = this.props.validateProcessState(this.state.spawner, this.isEdit);
        if (errMessage == null) {
            this.props.setProcessState(this.state.spawner);
        } else {
            return errMessage;
        }
        return null;
    }

    conditionForm = (spawnerSpec: SpawnerSpec): (JSX.Element) => {
        return <>
            <ConditionBuilder saveConditionHandler={this.saveConditionHandler()}
                              condition={spawnerSpec.condition}
                              enableButton={this.canAddCondition(spawnerSpec)} />
        </>
    }

    renderForm = (): (JSX.Element) => {
        const spawnerSpec = this.state.spawner.spawnerSpec;
        const spawnerKey = this.state.spawner.name;

        let conditionGroup = null;

        if (spawnerSpec.condition !== UndefinedCondition) {
            let condition = spawnerSpec.condition;
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
                <Form.Group controlId="spawnerName">
                    <Form.Label>Spawner Name</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={spawnerKey}/>
                    <Form.Text className="text-muted"> Unique Spawner name </Form.Text>
                </Form.Group>
                <Form.Group controlId="pathToExec">
                    <Form.Label>Path to Executable</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={spawnerSpec.pathToExec}/>
                    <Form.Text className="text-muted"> Absolute path to executable </Form.Text>
                </Form.Group>
                <Form.Group controlId="commandArgs">
                    <Form.Label>Command Args</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={spawnerSpec.commandArgs.join(",")}/>
                    <Form.Text className="text-muted"> Optional command arguments for executable </Form.Text>
                </Form.Group>
                <Form.Group controlId="delayInMs">
                    <Form.Label>Delay (in ms)</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} defaultValue={spawnerSpec.delayInMs}/>
                    <Form.Text className="text-muted"> Optional delay </Form.Text>
                </Form.Group>
                <Form.Group controlId="doSync">
                    <Form.Label>Synchronous</Form.Label>
                    <Form.Check
                        custom
                        type="checkbox"
                        defaultChecked={spawnerSpec.doSync}
                        onChange={this.onChange}
                    />
                    <Form.Text className="text-muted">Check to wait to command to complete</Form.Text>
                </Form.Group>
                {conditionGroup}
                <Form.Group>
                    {this.conditionForm(spawnerSpec)}
                </Form.Group>
            </Form>
        </>
    }

    render() {
        let title = "Add Spawner";
        if (this.isEdit) {
            title = "Update Spawner";
        }
        return <>
            <MainModal title={title} renderForm={this.renderForm} handleSave={this.handleSave}/>
        </>
    }
}

export class SpawnerSpec {
    condition: Condition
    delayInMs: number
    doSync: boolean
    pathToExec: string
    commandArgs: Array<string>

    constructor(condition: Condition, delayInMs: number, doSync: boolean,
                pathToExec: string, commandArgs: Array<string>) {
        this.delayInMs = delayInMs;
        this.doSync = doSync;
        this.pathToExec = pathToExec;
        this.commandArgs = commandArgs;

        if (condition) {
            this.condition = condition;
        } else {
            this.condition = UndefinedCondition;
        }
    }

    copy = (): SpawnerSpec => {
        if (this.condition !== UndefinedCondition) {
            return new SpawnerSpec(this.condition.copy(), this.delayInMs, this.doSync,
                this.pathToExec, this.commandArgs);
        } else {
            return new SpawnerSpec(this.condition, this.delayInMs, this.doSync,
                this.pathToExec, this.commandArgs);
        }
    }
}

export interface SpawnerJSONProps {
    name: string
    condition: ConditionJSONProps
    delayInMs: number
    doSync: boolean
    job: {
        runnable: {
            pathToExec: string
            cmdArgs: Array<string>
        }
    }
}

interface SpawnerProps {
    name: string
    spawnerSpec: SpawnerSpec
}

interface SpawnerState {
}

export class Spawner extends React.Component<SpawnerProps, SpawnerState> {
    name: string
    spawnerSpec: SpawnerSpec

    constructor(props: SpawnerProps) {
        super(props);
        this.name = props.name;
        this.spawnerSpec = props.spawnerSpec;
    }

    toString = ():string => {
        return JSON.stringify(this.toJSON());
    }

    static fromJSON = (props: SpawnerJSONProps): Spawner => {
        return new Spawner({
            name: props.name,
            spawnerSpec: new SpawnerSpec(
                Condition.fromJSON(props.condition),
                props.delayInMs,
                props.doSync,
                props.job.runnable.pathToExec,
                props.job.runnable.cmdArgs
            )
        })
    }

    toJSON = (): {} => {
        return {
            spawner: {
                name: this.name,
                condition: this.spawnerSpec.condition.toJSON(),
                delayInMs: this.spawnerSpec.delayInMs,
                doSync: this.spawnerSpec.doSync,
                job: {
                    runnable: {
                        pathToExec: this.spawnerSpec.pathToExec,
                        cmdArgs: this.spawnerSpec.commandArgs
                    }
                }
            }
        }
    }

    copy = ():Spawner => {
        return new Spawner({
            name: this.name,
            spawnerSpec: this.spawnerSpec
        })
    }

    render() {
        return <>
            {this.name}
        </>
    }
}