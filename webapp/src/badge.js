function isTwinPost(post) {
    return Boolean(post && post.props && post.props.twinmate && post.props.twinmate.generated === true);
}

module.exports = {isTwinPost};
