import * as React from "react";
import {ChangeEvent} from "react";
import {Button, Form, Modal} from "react-bootstrap";

export enum OperatorType {
    Unknown = "Unknown",
    Unary = "Unary",
    Binary = "Binary",
    Logical = "Logical",
    Comparator = "Comparator",
    Exists = "Exists",
}

export function StrToOperatorType(strOperatorType: string): OperatorType {
    return OperatorType[strOperatorType as keyof typeof OperatorType];
}

export function OperatorTypeToStr(operatorType: OperatorType): string {
    return operatorType
}

export function GetOperatorTypes(): Array<string> {
    return Object.keys(OperatorType).filter((v) => v !== "Unknown")
}

export function HasLHS(operatorType: OperatorType): boolean {
    if (!operatorType) {
        return false;
    }
    return operatorType !== OperatorType.Unary && operatorType !== OperatorType.Exists;

}

export function HasRHS(operatorType: OperatorType): boolean {
    if (!operatorType) {
        return false;
    }
    return true
}

export enum ExistsOperator {
    Exists = "Exists",
    NotExists = "NotExists"
}

export function StrToExistsOperator(strExistsOperator: string): ExistsOperator {
    return ExistsOperator[strExistsOperator as keyof typeof ExistsOperator];
}

export function ExistsOperatorToStr(operator: ExistsOperator): string {
    return operator
}

export function GetExistsOperators(): Array<string> {
    return Object.keys(ExistsOperator).filter((v) => v !== "Unknown")
}

export enum UnaryOperator {
    Negation = "Negation",
    Inversion = "Inversion",
    LogicalNot = "LogicalNot"
}

export function StrToUnaryOperator(strUnaryOperator: string): UnaryOperator {
    return UnaryOperator[strUnaryOperator as keyof typeof UnaryOperator];
}

export function UnaryOperatorToStr(operator: UnaryOperator): string {
    return operator
}

export function GetUnaryOperators(): Array<string> {
    return Object.keys(UnaryOperator).filter((v) => v !== "Unknown")
}

export enum BinaryOperator {
    Addition = "Addition",
    Subtract = "Subtract",
    Multiply = "Multiply",
    Divide = "Divide",
    Power = "Power",
    Modulus = "Modulus",
    RightShift = "RightShift",
    LeftShift = "LeftShift",
    Or = "Or",
    And = "And",
    Xor = "Xor"
}

export function StrToBinaryOperator(strBinaryOperator: string): BinaryOperator {
    return BinaryOperator[strBinaryOperator as keyof typeof BinaryOperator];
}

export function BinaryOperatorToStr(operator: BinaryOperator): string {
    return operator
}

export function GetBinaryOperators(): Array<string> {
    return Object.keys(BinaryOperator).filter((v) => v !== "Unknown")
}

export enum LogicalOperator {
    And = "And",
    Or = "Or"
}

export function StrToLogicalOperator(strLogicalOperator: string): LogicalOperator {
    return LogicalOperator[strLogicalOperator as keyof typeof LogicalOperator];
}

export function LogicalOperatorToStr(operator: LogicalOperator): string {
    return operator
}

export function GetLogicalOperators(): Array<string> {
    return Object.keys(LogicalOperator).filter((v) => v !== "Unknown")
}

export enum ComparatorOperator {
    GreaterThan = "GreaterThan",
    LessThan = "LessThan",
    GreaterThanOrEqual = "GreaterThanOrEqual",
    LessThanOrEqual = "LessThanOrEqual",
    Equal = "Equal",
    NotEqual = "NotEqual",
    RegexMatch = "RegexMatch",
    RegexNoMatch = "RegexNoMatch",
}

export function StrToComparatorOperator(strComparatorOperator: string): ComparatorOperator {
    return ComparatorOperator[strComparatorOperator as keyof typeof ComparatorOperator];
}

export function ComparatorOperatorToStr(operator: ComparatorOperator): string {
    return operator
}

export function GetComparatorOperators(): Array<string> {
    return Object.keys(ComparatorOperator).filter((v) => v !== "Unknown")
}

export type Operator = ExistsOperator | UnaryOperator | BinaryOperator | LogicalOperator | ComparatorOperator | null;

function StrToOperator(strOperator: string, operatorType: OperatorType): Operator {
    if (operatorType === OperatorType.Exists) {
        return StrToExistsOperator(strOperator);
    } else if (operatorType === OperatorType.Unary) {
        return StrToUnaryOperator(strOperator);
    } else if (operatorType === OperatorType.Binary) {
        return StrToBinaryOperator(strOperator);
    } else if (operatorType === OperatorType.Logical) {
        return StrToLogicalOperator(strOperator);
    } else if (operatorType === OperatorType.Comparator) {
        return StrToComparatorOperator(strOperator);
    }
    return null;
}

function OperatorToStr(operator: Operator): string {
    if (operator) {
        return operator.toString();
    }
    return "";
}

interface ConditionBuilderProps {
    condition?: Condition
    saveConditionHandler: SetConditionFn
    enableButton: boolean
}

interface ConditionBuilderState {
    condition: Condition
    showModal: boolean
}

export class ConditionBuilder extends React.Component<ConditionBuilderProps, ConditionBuilderState> {
    isEdit: boolean

    constructor(props: ConditionBuilderProps) {
        super(props);
        let condition: Condition | undefined = this.props.condition;
        this.isEdit = true;
        if (this.props.condition == null) {
            this.isEdit = false;
            condition = new Condition({
                operatorType: OperatorType.Exists,
                operator: null,
                lhs: "",
                rhs: "",
            })
        }
        if (condition) {
            this.state = {
                condition: condition,
                showModal: false,
            }
        }
    }

    handleShowModal = () => {
        this.setState ({
            showModal: true
        });
    }

    handleHideModal = () => {
        this.setState ({
            showModal: false
        });
    }

    componentDidMount() {
    }

    onChange = (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        let condition = this.state.condition.copy();
        let operatorType = condition.operatorType;
        if (event.target.id === "operatorType"){
            condition.operatorType = StrToOperatorType(event.target.value);
        } else if (event.target.id === "operator"){
            if (operatorType === OperatorType.Exists) {
                condition.operator = StrToExistsOperator(event.target.value);
            } else if (operatorType === OperatorType.Unary) {
                condition.operator = StrToUnaryOperator(event.target.value);
            } else if (operatorType === OperatorType.Binary) {
                condition.operator = StrToBinaryOperator(event.target.value);
            } else if (operatorType === OperatorType.Logical) {
                condition.operator = StrToLogicalOperator(event.target.value);
            } else if (operatorType === OperatorType.Comparator) {
                condition.operator = StrToComparatorOperator(event.target.value);
            }
        } else if (event.target.id === "lhs") {
            condition.lhs = event.target.value;
        } else if (event.target.id === "rhs") {
            condition.rhs = event.target.value;
        }

        this.setState({
            condition: condition
        })
    }

    handleSave = (): any => {
        this.props.saveConditionHandler(this.state.condition);
        this.handleHideModal();
    }

    renderForm = (): (JSX.Element) => {
        let operatorType = this.state.condition.operatorType;
        const operator = this.state.condition.operator;
        const lhs = this.state.condition.lhs;
        const rhs = this.state.condition.rhs;
        let operators: Array<string> = [];
        let defaultChoice = "Choose...";
        let defaultOperatorType = defaultChoice;
        let defaultOperator = defaultChoice;

        if (operatorType) {
            defaultOperatorType = OperatorTypeToStr(operatorType);
        }

        if (operator) {
            defaultOperator = OperatorToStr(operator);
        }

        if (operatorType === OperatorType.Exists) {
            operators = GetExistsOperators();
        } else if (operatorType === OperatorType.Unary) {
            operators = GetUnaryOperators();
        } else if (operatorType === OperatorType.Binary) {
            operators = GetBinaryOperators();
        } else if (operatorType === OperatorType.Logical) {
            operators = GetLogicalOperators();
        } else if (operatorType === OperatorType.Comparator) {
            operators = GetComparatorOperators();
        }

        return <>
            <Form>
                <Form.Group controlId="operatorType">
                    <Form.Label>Condition Operator Type</Form.Label>
                    <Form.Control as="select" onChange={this.onChange} defaultValue={defaultOperatorType}>
                        <option key={defaultChoice}>{defaultChoice}</option>
                        {GetOperatorTypes().map(t => <option key={t}>{t}</option>)}
                    </Form.Control>
                    <Form.Text className="text-muted">
                        Choose an operator type
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId="lhs">
                    <Form.Label>LHS</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} value={lhs}
                                  disabled={!HasLHS(operatorType)}/>
                    <Form.Text className="text-muted">
                        Left-hand side of expression
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId="operator">
                    <Form.Label>Operator</Form.Label>
                    <Form.Control as="select" onChange={this.onChange} defaultValue={defaultOperator}
                                  disabled={!operatorType}>
                        <option key={defaultChoice}>{defaultChoice}</option>
                        {operators.map(t => <option key={t}>{t}</option>)}
                    </Form.Control>
                    <Form.Text className="text-muted">
                        Choose an operator
                    </Form.Text>
                </Form.Group>
                <Form.Group controlId="rhs">
                    <Form.Label>RHS</Form.Label>
                    <Form.Control type="text" onChange={this.onChange} value={rhs} disabled={!HasRHS(operatorType)}/>
                    <Form.Text className="text-muted">
                        Right-hand side of expression
                    </Form.Text>
                </Form.Group>
            </Form>
        </>
    }

    render() {
        return <>
            <Button variant="primary" onClick={this.handleShowModal} disabled={!this.props.enableButton}>
                Add Condition
            </Button>
            <Modal show={this.state.showModal} onHide={this.handleHideModal}>
                <Modal.Header closeButton>
                    <Modal.Title>Add Condition</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    {this.renderForm()}
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={this.handleHideModal}>
                        Close
                    </Button>
                    <Button variant="primary" onClick={this.handleSave}>
                        Save Changes
                    </Button>
                </Modal.Footer>
            </Modal>
        </>
    }
}


interface ConditionProps {
    operatorType: OperatorType
    operator: Operator
    lhs: string
    rhs: string
}

interface ConditionState {}

function operandIsVariable(operand: string): boolean {
    if (operand.length === 0) {
        return false;
    }
    return operand[0] === '[' && operand[operand.length - 1] === ']';

}

function operandIsNumeric(operand: string): boolean {
    return !isNaN(Number(operand));
}

function operandSpecToString(operandSpec: {}): string {
    if ("variable" in operandSpec) {
        return '[' + operandSpec["variable"]["name"] as string + ']';
    } else if ("numeric" in operandSpec) {
        return operandSpec["numeric"] as string;
    } else if ("literal" in operandSpec) {
        return operandSpec["literal"] as string;
    }
    return ""
}

function getOperandSpec(operand: string): {} {
    if (operandIsVariable(operand)) {
        return {
            variable: {
                name: operand.replace("[", "").replace("]", "")
            }
        }
    } else if (operandIsNumeric(operand)) {
        return {
            numeric: operand
        }
    }
    return {
        literal: operand
    }
}

interface ComparatorJSONProps {
    discriminator: 'ComparatorJSONProps';
    expression: {
        comparator: {
            lhs: {}
            rhs: {}
            op: string
        }
    }
}

interface LogicalJSONProps {
    discriminator: 'LogicalJSONProps';
    expression: {
        logical: {
            lhs: {}
            rhs: {}
            op: string
        }
    }
}

interface BinaryJSONProps {
    discriminator: 'BinaryJSONProps';
    expression: {
        binary: {
            lhs: {}
            rhs: {}
            op: string
        }
    }
}

interface UnaryJSONProps {
    discriminator: 'UnaryJSONProps';
    expression: {
        unary: {
            rhs: {}
            op: string
        }
    }
}

type ExpressionJSONProps = ComparatorJSONProps | LogicalJSONProps | BinaryJSONProps | UnaryJSONProps;

interface ExistsJSONProps {
    discriminator: 'ExistsJSONProps';
    exists: {
        ops: [
            {
                key: string
                op: string
            }
        ]
    }
}

export type ConditionJSONProps = ExpressionJSONProps | ExistsJSONProps;

export class Condition extends React.Component<ConditionProps, ConditionState> {
    operatorType: OperatorType
    operator: Operator
    lhs: string
    rhs: string

    constructor(props: ConditionProps) {
        super(props);
        this.operatorType = props.operatorType;
        this.operator = props.operator;
        this.lhs = props.lhs;
        this.rhs = props.rhs;
    }

    copy = ():Condition => {
        return new Condition({
            operatorType: this.operatorType,
            operator: this.operator,
            lhs: this.lhs,
            rhs: this.rhs,
        })
    }

    static fromJSON = (props: ConditionJSONProps): Condition => {
        if (!props) {
            return UndefinedCondition;
        }
        if ("expression" in props && "comparator" in props.expression) {
            return new Condition({
                operatorType: OperatorType.Comparator,
                operator: StrToOperator(props.expression.comparator.op, OperatorType.Comparator),
                lhs: operandSpecToString(props.expression.comparator.lhs),
                rhs: operandSpecToString(props.expression.comparator.rhs)
            })
        } else if ("expression" in props && "logical" in props.expression) {
            return new Condition({
                operatorType: OperatorType.Logical,
                operator: StrToOperator(props.expression.logical.op, OperatorType.Logical),
                lhs: operandSpecToString(props.expression.logical.lhs),
                rhs: operandSpecToString(props.expression.logical.rhs)
            })
        } else if ("expression" in props && "binary" in props.expression) {
            return new Condition({
                operatorType: OperatorType.Binary,
                operator: StrToOperator(props.expression.binary.op, OperatorType.Binary),
                lhs: operandSpecToString(props.expression.binary.lhs),
                rhs: operandSpecToString(props.expression.binary.rhs)
            })
        } else if ("expression" in props && "unary" in props.expression) {
            return new Condition({
                operatorType: OperatorType.Unary,
                operator: StrToOperator(props.expression.unary.op, OperatorType.Unary),
                lhs: "",
                rhs: operandSpecToString(props.expression.unary.rhs)
            })
        } else if ("exists" in props && "ops" in props.exists) {
            return new Condition({
                operatorType: OperatorType.Exists,
                operator: StrToOperator(props.exists.ops[0].op, OperatorType.Exists),
                lhs: "",
                rhs: props.exists.ops[0].key
            })
        }
        return UndefinedCondition;
    }

    toJSON = (): {} => {
        if (this.operatorType === OperatorType.Comparator) {
            return {
                expression: {
                    comparator: {
                        lhs: getOperandSpec(this.lhs),
                        rhs: getOperandSpec(this.rhs),
                        op: this.operator
                    }
                }
            }
        } else if (this.operatorType === OperatorType.Logical) {
            return {
                expression: {
                    logical: {
                        lhs: getOperandSpec(this.lhs),
                        rhs: getOperandSpec(this.rhs),
                        op: this.operator
                    }
                }
            }
        } else if (this.operatorType === OperatorType.Binary) {
            return {
                expression: {
                    binary: {
                        lhs: getOperandSpec(this.lhs),
                        rhs: getOperandSpec(this.rhs),
                        op: this.operator
                    }
                }
            }
        } else if (this.operatorType === OperatorType.Unary) {
            return {
                expression: {
                    unary: {
                        rhs: getOperandSpec(this.rhs),
                        op: this.operator
                    }
                }
            }
        } else if (this.operatorType === OperatorType.Exists) {
            return {
                exists: {
                    ops: [
                        {
                            key: this.rhs,
                            op: this.operator
                        }
                    ]
                }
            }
        }

        return {}
    }

    toString = ():string => {
        if (this.lhs) {
            return this.lhs + ' ' + this.operator + ' ' + this.rhs;
        }
        return this.operator + ' ' + this.rhs;
    }

    render() {
        return <>
            {this.toString()}
        </>
    }
}

export const UndefinedCondition = new Condition({
    operatorType: OperatorType.Unknown,
    operator: null,
    lhs: "",
    rhs: ""
});

export type SetConditionFn = (condition: Condition) => void;
