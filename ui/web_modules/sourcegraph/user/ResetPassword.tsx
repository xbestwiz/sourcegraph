// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Link} from "react-router";
import Helmet from "react-helmet";
import * as classNames from "classnames";

import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Button, Input} from "sourcegraph/components/index";

import * as UserActions from "sourcegraph/user/UserActions";
import {UserStore} from "sourcegraph/user/UserStore";

import "sourcegraph/user/UserBackend"; // for side effects
import {redirectIfLoggedIn} from "sourcegraph/user/redirectIfLoggedIn";
import * as styles from "sourcegraph/user/styles/accountForm.css";

type Props = {
	location?: any,
};

class ResetPasswordComp extends Container<Props, any> {
	static contextTypes = {
		user: React.PropTypes.object,
	};

	_passwordInput: any;
	_confirmInput: any;

	constructor(props: Props) {
		super(props);
		this._passwordInput = null;
		this._confirmInput = null;
		this._handleSubmit = this._handleSubmit.bind(this);
	}

	reconcileState(state, props, context) {
		Object.assign(state, props);
		state.token = state.location.query && state.location.query.token; // TODO: error handling (missing token)
		state.pendingAuthAction = UserStore.pendingAuthActions["reset"] || false;
		state.authResponse = UserStore.authResponses["reset"] || null;
	}

	stores() { return [UserStore]; }

	_handleSubmit(ev) {
		ev.preventDefault();
		let action = new UserActions.SubmitResetPassword(
			this._passwordInput.value,
			this._confirmInput.value,
			this.state.token
		);
		Dispatcher.Stores.dispatch(action);
		Dispatcher.Backends.dispatch(action);
	}

	render(): JSX.Element | null {
		return (
			<form className={classNames(styles.full_page, styles.form)} onSubmit={this._handleSubmit}>
				<Helmet title="Reset Password" />
				<h1>Reset your password</h1>
				<label>
					<span>New password</span>
					<Input type="password"
						domRef={(e) => this._passwordInput = e}
						name="password"
						autoComplete="new-password"
						autoFocus={true}
						block={true}
						required={true} />
				</label>
				<label>
					<span>Confirm new password</span>
					<Input type="password"
						domRef={(e) => this._confirmInput = e}
						block={true}
						required={true} />
				</label>
				<Button color="blue"
					block={true}
					loading={this.state.pendingAuthAction}>Reset Password</Button>
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Error &&
					<div className={styles.error}>{this.state.authResponse.Error.body.message}</div>
				}
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Success &&
					<div className={styles.success}>
						Your password has been reset! <Link to="/login">Sign in.</Link>
					</div>
				}
			</form>
		);
	}
}

export const ResetPassword = redirectIfLoggedIn("/", ResetPasswordComp);
